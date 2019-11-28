package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/agent"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/oklog/ulid"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/namespace"
)

var (
	_ agent.Repository        = (*Repository)(nil)
	_ controlplane.Repository = (*Repository)(nil)

	minULID = ulid.MustNew(0, zeroReader{})
	maxULID = ulid.MustNew(ulid.MaxTime(), oneReader{})
)

const (
	runsPrefix   = "runs/"
	statesPrefix = "states/"
	agentsPrefix = "agents/"
	nodesPrefix  = "nodes/"
)

// Repository is the etcd backed implementation of an adagio Repository type (control plane and agent)
// It facilitates a distributed set of agents and control plane consumers using etcd as the store
// and etcd v3 transactions to ensure a consistent behavior
// It adheres to the repository test harness
type Repository struct {
	kv      clientv3.KV
	watcher clientv3.Watcher
	leaser  clientv3.Lease

	mu            sync.Mutex
	subscriptions map[chan<- *adagio.Event]chan struct{}

	namespace string
	list      string
	now       func() time.Time

	ttl     time.Duration
	leases  map[string]func()
	leaseMu sync.Mutex
}

// New constructs and configure a new repository service from the provided etcd client
// and a set of function options
func New(kv clientv3.KV, watcher clientv3.Watcher, leaser clientv3.Lease, opts ...Option) *Repository {
	r := &Repository{
		kv:            kv,
		watcher:       watcher,
		leaser:        leaser,
		subscriptions: map[chan<- *adagio.Event]chan struct{}{},
		now:           func() time.Time { return time.Now().UTC() },
		namespace:     "v0",
		list:          "default",
		ttl:           10 * time.Second,
		leases:        map[string]func(){},
	}

	Options(opts).Apply(r)

	fullNS := path.Join(r.namespace, r.list) + "/"

	r.kv = namespace.NewKV(r.kv, fullNS)
	r.watcher = namespace.NewWatcher(r.watcher, fullNS)
	r.leaser = namespace.NewLease(r.leaser, fullNS)

	return r
}

// Stats returns the state of world represented as counts within the etcd database
func (r *Repository) Stats(ctx context.Context) (*adagio.Stats, error) {
	stats := &adagio.Stats{
		NodeCounts: &adagio.Stats_NodeCounts{},
	}

	resp, err := r.kv.Get(ctx,
		runsPrefix,
		clientv3.WithPrefix(),
		clientv3.WithCountOnly())
	if err != nil {
		return nil, err
	}

	stats.RunCount = resp.Count

	for status := range adagio.Node_Status_name {
		resp, err := r.kv.Get(ctx,
			nodesInStateKey(adagio.Node_Status(status)),
			clientv3.WithPrefix(),
			clientv3.WithCountOnly(),
			clientv3.WithRev(resp.Header.Revision))
		if err != nil {
			return nil, err
		}

		switch adagio.Node_Status(status) {
		case adagio.Node_WAITING:
			stats.NodeCounts.WaitingCount = resp.Count
		case adagio.Node_READY:
			stats.NodeCounts.ReadyCount = resp.Count
		case adagio.Node_RUNNING:
			stats.NodeCounts.RunningCount = resp.Count
		case adagio.Node_COMPLETED:
			stats.NodeCounts.CompletedCount = resp.Count
		}
	}

	return stats, nil
}

// StartRun takes a graph specification and instantiates it within etcd an returns the resulting Run
// representation
func (r *Repository) StartRun(ctx context.Context, spec *adagio.GraphSpec) (run *adagio.Run, err error) {
	run, err = adagio.NewRun(spec)
	if err != nil {
		return
	}

	data, err := marshalRun(run.CreatedAt, spec.Nodes, run.Edges)
	if err != nil {
		return nil, err
	}

	var (
		runKey = runKey(run)
		cmps   = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(runKey), "=", 0),
		}
		ops = []clientv3.Op{
			clientv3.OpPut(runKey, string(data)),
		}
	)

	for _, node := range run.Nodes {
		nodeData, err := json.Marshal(node)
		if err != nil {
			return nil, err
		}

		var (
			key      = nodeKey(run.Id, node.Spec.Name)
			stateKey = nodeInStateKey(run.Id, statusToString(node.Status), node.Spec.Name)
			put      = clientv3.OpPut(key, string(nodeData))
			putState = clientv3.OpPut(stateKey, "")
		)

		ops = append(ops, put, putState)
	}

	resp, err := r.kv.Txn(ctx).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		return
	}

	if !resp.Succeeded {
		err = errors.New("duplicate run already created")
	}

	return
}

// InspectRun takes an ID and returns the associated Run if found within etcd
func (r *Repository) InspectRun(ctx context.Context, id string) (*adagio.Run, error) {
	return r.getRun(ctx, id)
}

// ListAgents returns at agents recorded within etcd at the time of the call
func (r *Repository) ListAgents(ctx context.Context) (agents []*adagio.Agent, err error) {
	resp, err := r.kv.Get(ctx, agentsPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		var agent adagio.Agent
		if err := json.Unmarshal(kv.Value, &agent); err != nil {
			return nil, err
		}

		agents = append(agents, &agent)
	}

	return
}

// ListRuns returns a list of runs given a set of predicates
// Given no start time is provided now is assumed
// Given no finish time is provided epoch is assumed
// Given no limit is provided all runs are returned
func (r *Repository) ListRuns(ctx context.Context, req controlplane.ListRequest) (runs []*adagio.Run, err error) {
	var (
		opts = []clientv3.OpOption{clientv3.WithPrefix()}
		key  = runsPrefix
	)

	if req.Start != nil || req.Finish != nil {
		var (
			start  = maxULID
			finish = minULID
		)

		if req.Start != nil {
			start, err = ulid.New(ulid.Timestamp(*req.Start), oneReader{})
			if err != nil {
				return
			}
		}

		if req.Finish != nil {
			finish, err = ulid.New(ulid.Timestamp(*req.Finish), zeroReader{})
			if err != nil {
				return
			}
		}

		key = runsPrefix + finish.String()
		opts = []clientv3.OpOption{clientv3.WithRange(runsPrefix + start.String())}
	}

	if req.Limit != nil {
		opts = append(opts, clientv3.WithLimit(int64(*req.Limit)))
	}

	resp, err := r.kv.Get(ctx,
		key,
		append(opts, clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))...)
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		// strip list start key
		parts := strings.Split(string(kv.Key), "/")

		// ignore non-run keys
		if len(parts) != 2 {
			continue
		}

		run, err := r.getRun(ctx, parts[1])
		if err != nil {
			return nil, err
		}

		runs = append(runs, run)
	}

	return
}

// ClaimNode attempts to claim a node identified by name for a specified run ID and providing a unique claim
// Given the node is found and the claim is successful the node is returned and the claimed boolean with be true
func (r *Repository) ClaimNode(ctx context.Context, runID, name string, claim *adagio.Claim) (node *adagio.Node, claimed bool, err error) {
	run, err := r.getRun(ctx, runID)
	if err != nil {
		return nil, false, err
	}

	node, err = run.GetNodeByName(name)
	if err != nil {
		return
	}

	if node.Status == adagio.Node_WAITING {
		return nil, false, adagio.ErrNodeNotReady
	}

	// node must be either ready or in none state
	if node.Status != adagio.Node_READY && node.Status != adagio.Node_NONE {
		return nil, false, nil
	}

	// construct a lease which is kept-alive
	leaseID, err := r.lease(claim.Id)
	if err != nil {
		return nil, false, err
	}

	// set claim on node
	node.Claim = claim

	defer func() {
		if !claimed {
			r.cancelLease(claim.Id)
		}
	}()

	cmps, ops, err := r.transition(ctx, runID, node, adagio.Node_RUNNING, nil, nil, clientv3.WithLease(leaseID))
	if err != nil {
		return nil, false, err
	}

	resp, err := r.kv.Txn(ctx).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		return nil, false, err
	}

	if !resp.Succeeded {
		return nil, false, nil
	}

	return node, resp.Succeeded, nil
}

func (r *Repository) complete(ctx context.Context, runID string, node *adagio.Node, cmps []clientv3.Cmp, ops []clientv3.Op) ([]clientv3.Cmp, []clientv3.Op, error) {
	// given node has not already been completed
	if node.Status != adagio.Node_COMPLETED {
		var err error
		cmps, ops, err = r.transition(ctx, runID, node, adagio.Node_COMPLETED, cmps, ops)
		if err != nil {
			return nil, nil, err
		}
	}

	return cmps, ops, nil
}

func (r *Repository) transition(ctx context.Context, runID string, node *adagio.Node, toStatus adagio.Node_Status, cmps []clientv3.Cmp, ops []clientv3.Op, putOpts ...clientv3.OpOption) ([]clientv3.Cmp, []clientv3.Op, error) {
	var (
		nodeKey = nodeKey(runID, node.Spec.Name)
		fromKey = nodeInStateKey(runID, statusToString(node.Status), node.Spec.Name)
		toKey   = nodeInStateKey(runID, statusToString(toStatus), node.Spec.Name)
		from    = node.Status
	)

	node.Status = toStatus

	switch toStatus {
	case adagio.Node_RUNNING:
		node.StartedAt = r.now().Format(time.RFC3339Nano)

	case adagio.Node_COMPLETED:
		now := r.now()
		if node.StartedAt == "" {
			node.StartedAt = now.Format(time.RFC3339Nano)
		}

		node.FinishedAt = now.Format(time.RFC3339Nano)
	}

	data, err := json.Marshal(node)
	if err != nil {
		return nil, nil, err
	}

	if from != adagio.Node_NONE {
		return append(cmps,
				// ensure node exists
				clientv3.Compare(clientv3.Version(nodeKey), ">", 0),
				// ensure node has not moved into to state yet
				clientv3.Compare(clientv3.Version(toKey), "=", 0),
				// ensure node starts in expected state
				clientv3.Compare(clientv3.Version(fromKey), ">", 0)),
			append(ops,
				clientv3.OpPut(nodeKey, string(data)),
				clientv3.OpDelete(fromKey),
				clientv3.OpPut(toKey, "", putOpts...),
			), nil
	}

	return append(cmps,
			append(
				// ensure node does not exist in any state
				r.nodeIsOrphaned(runID, node),
				// ensure node exists
				clientv3.Compare(clientv3.Version(nodeKey), ">", 0),
			)...),
		append(ops,
			clientv3.OpPut(nodeKey, string(data)),
			clientv3.OpPut(toKey, "", putOpts...),
		), nil
}

func (r *Repository) nodeIsOrphaned(runID string, node *adagio.Node) []clientv3.Cmp {
	return []clientv3.Cmp{
		// ensure node does not exist in any state
		r.statusDoesNotExist(runID, node, adagio.Node_WAITING),
		r.statusDoesNotExist(runID, node, adagio.Node_READY),
		r.statusDoesNotExist(runID, node, adagio.Node_RUNNING),
		r.statusDoesNotExist(runID, node, adagio.Node_COMPLETED),
	}
}

func (r *Repository) statusDoesNotExist(runID string, node *adagio.Node, status adagio.Node_Status) clientv3.Cmp {
	statusKey := nodeInStateKey(runID, statusToString(status), node.Spec.Name)
	return clientv3.Compare(clientv3.Version(statusKey), "=", 0)
}

func (r *Repository) lease(claimID string) (clientv3.LeaseID, error) {
	// store lease keep-alive cancel func
	ctx, cancel := context.WithCancel(context.Background())

	// grant lease in seconds
	leaseResp, err := r.leaser.Grant(ctx, int64(r.ttl/time.Second))
	if err != nil {
		return 0, err
	}

	r.leaseMu.Lock()
	r.leases[claimID] = func() {
		// cancel keep-alive
		cancel()

		// revoke lease
		if _, err := r.leaser.Revoke(context.Background(), leaseResp.ID); err != nil {
			log.Println(claimID, err)
			return
		}
	}
	r.leaseMu.Unlock()

	// keep lease alive
	go func() {
		defer r.cancelLease(claimID)

		resps, err := r.leaser.KeepAlive(ctx, leaseResp.ID)
		if err != nil {
			log.Println(err)
			return
		}

		for resp := range resps {
			log.Println("keep-alive", resp.ID)
		}

		log.Println("finished keep-alive for", claimID)
	}()

	return leaseResp.ID, nil
}

func (r *Repository) cancelLease(claimID string) {
	// clear lease keep-alive
	r.leaseMu.Lock()
	if cancel, ok := r.leases[claimID]; ok {
		cancel()
		delete(r.leases, claimID)
	}
	r.leaseMu.Unlock()
}

// FinishNode records a result for a node identified by name for a specified run ID and given a unique and active claim
func (r *Repository) FinishNode(ctx context.Context, runID, name string, result *adagio.Node_Result, claim *adagio.Claim) error {
	run, err := r.getRun(ctx, runID)
	if err != nil {
		return err
	}

	node, err := run.GetNodeByName(name)
	if err != nil {
		return err
	}

	if node.Status != adagio.Node_RUNNING {
		return errors.New("attempt to finish non-running node")
	}

	// append result to list of attempts
	node.Attempts = append(node.Attempts, result)

	var (
		cmps []clientv3.Cmp
		ops  []clientv3.Op
	)

	if result.Conclusion == adagio.Node_Result_SUCCESS {
		cmps, ops, err = r.handleSuccess(ctx, run, node, result)
		if err != nil {
			return err
		}
	} else {
		cmps, ops, err = r.handleFailure(ctx, run, node, result)
		if err != nil {
			return err
		}
	}

	resp, err := r.kv.Txn(ctx).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		r.cancelLease(claim.Id)

		return err
	}

	if !resp.Succeeded {
		return r.FinishNode(ctx, run.Id, node.Spec.Name, result, claim)
	}

	r.cancelLease(claim.Id)

	return nil
}

func (r *Repository) handleSuccess(ctx context.Context, run *adagio.Run, node *adagio.Node, result *adagio.Node_Result) ([]clientv3.Cmp, []clientv3.Op, error) {
	cmps, ops, err := r.complete(ctx, run.Id, node, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	graph := adagio.GraphFrom(run)

	outgoing, err := graph.Outgoing(node)
	if err != nil {
		return nil, nil, err
	}

	for o := range outgoing {
		out := o.(*adagio.Node)

		if out.Status > adagio.Node_WAITING {
			continue
		}

		isReady := true

		incoming, err := graph.Incoming(out)
		if err != nil {
			return nil, nil, err
		}

		for v := range incoming {
			in := v.(*adagio.Node)

			// target node is ready if all incoming nodes
			// are completed
			isReady = isReady && in.Status == adagio.Node_COMPLETED

			if in == node {
				// we have already considered the finishing node
				continue
			}

			if in.Status == adagio.Node_NONE {
				cmps = append(cmps, r.nodeIsOrphaned(run.Id, in)...)
				continue
			}

			// ensure node in state key is created
			currentKey := nodeInStateKey(run.Id, statusToString(in.Status), in.Spec.Name)
			cmps = append(cmps, clientv3.Compare(clientv3.Version(currentKey), ">", 0))
		}

		// if all nodes are now completed
		// then the outgoing target is ready
		if isReady {
			cmps, ops, err = r.transition(ctx, run.Id, out, adagio.Node_READY, cmps, ops)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return cmps, ops, nil
}

func (r *Repository) handleFailure(ctx context.Context, run *adagio.Run, node *adagio.Node, result *adagio.Node_Result) ([]clientv3.Cmp, []clientv3.Op, error) {
	if adagio.CanRetry(node) {
		// put node back into the ready state to be attempted again
		return r.transition(ctx, run.Id, node, adagio.Node_READY, nil, nil)
	}

	cmps, ops, err := r.complete(ctx, run.Id, node, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	if err := adagio.GraphFrom(run).WalkFrom(node, func(gnode graph.Node) error {
		var (
			node, _ = gnode.(*adagio.Node)
			err     error
		)

		// complete outgoing nodes with nil result to signify
		// no attempt has been made
		cmps, ops, err = r.complete(ctx, run.Id, node, cmps, ops)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return cmps, ops, nil
}

// Subscribe registers the agent as a subscriber and sends events regarding node readiness and orphanage
// on the provided events channel
func (r *Repository) Subscribe(ctx context.Context, a *adagio.Agent, events chan<- *adagio.Event, typ ...adagio.Event_Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// create agent record

	// construct a lease for the agent
	leaseID, err := r.lease(a.Id)
	if err != nil {
		return err
	}

	// persist agent in keyspace
	agentData, err := json.Marshal(a)
	if err != nil {
		return err
	}

	if _, err := r.kv.Put(ctx, agentKey(a), string(agentData), clientv3.WithLease(leaseID)); err != nil {
		return err
	}

	// begin subscription

	ch := make(chan struct{})

	r.subscriptions[events] = ch

	go func() {
		defer r.cancelLease(a.Id)

		var (
			// construct a new context for this operation
			// as the subscription will be cancelled via
			// an unsubscribe
			ctx    = context.Background()
			opts   = []clientv3.OpOption{clientv3.WithPrefix()}
			filter = filter{
				orphaned: !types(typ).contains(adagio.Event_NODE_ORPHANED),
				ready:    !types(typ).contains(adagio.Event_NODE_READY),
			}
		)

		// consume existing ready nodes
		resp, err := r.kv.Get(ctx, nodesInStateKey(adagio.Node_READY), clientv3.WithPrefix())
		if err != nil {
			log.Println(err)
			goto Watch
		}

		// handle ready node events
		for _, kv := range resp.Kvs {
			r.handleKeyEvent(ctx, events, keyEvent{kv.Key, keyCreated}, filter, clientv3.WithRev(resp.Header.Revision))
		}

		// set the watch responses to return a revision higher than the response
		// to reduce the chance we observe the same ready node twice
		opts = append(opts, clientv3.WithRev(resp.Header.Revision+1))

	Watch:
		// watch for new events
		watch := r.watcher.Watch(ctx, statesPrefix, opts...)
		for {
			var resp clientv3.WatchResponse
			select {
			case <-ch:
				return
			case resp = <-watch:
			}

			if resp.Err() != nil {
				log.Println(resp.Err())
				continue
			}

			for _, ev := range resp.Events {
				kev := keyEvent{
					Key: ev.Kv.Key,
				}

				if ev.IsCreate() {
					kev.Type = keyCreated
				}

				// todo(george): import issues with etcd returning old urls from types
				// defined in new url scheme makes importing other etcd v3 libs to break
				// until then I am just going to compare to the literal 1 value which means
				// DELETE operation
				if ev.Type == 1 {
					kev.Type = keyDeleted
				}

				r.handleKeyEvent(ctx, events, kev, filter, clientv3.WithRev(resp.Header.Revision))
			}
		}
	}()

	return nil
}

type keyEvent struct {
	Key  []byte
	Type keyEventType
}

type keyEventType int

const (
	keyUnknown keyEventType = iota
	keyCreated
	keyDeleted
)

type filter struct {
	orphaned bool
	ready    bool
}

func (r *Repository) handleKeyEvent(ctx context.Context, dest chan<- *adagio.Event, ev keyEvent, filter filter, opts ...clientv3.OpOption) {
	keyParts := strings.Split(string(ev.Key), "/")
	if len(keyParts) < 6 {
		return
	}

	status := stringToStatus(keyParts[1])
	switch status {
	// we're only interested in actions on ready and running keys
	case adagio.Node_READY, adagio.Node_RUNNING:
	default:
		return
	}

	run, err := r.getRun(ctx, keyParts[3], opts...)
	if err != nil {
		log.Println(keyParts[3], err)
		return
	}

	node, err := run.GetNodeByName(keyParts[5])
	if err != nil {
		log.Println(err)
		return
	}

	switch ev.Type {
	case keyCreated:
		// if a ready status key has been created and the subscription contains a
		// node ready type then send a node ready event
		if status == adagio.Node_READY && !filter.ready {
			dest <- &adagio.Event{
				Type:     adagio.Event_NODE_READY,
				RunID:    keyParts[3],
				NodeSpec: node.Spec,
			}
		}
	case keyDeleted:
		// given the deletion of a running key where no other state key for the
		// node exists (this is where GetNodeByName returns a node with a NONE status)
		if status == adagio.Node_RUNNING && node.Status == adagio.Node_NONE && !filter.orphaned {
			dest <- &adagio.Event{
				Type:     adagio.Event_NODE_ORPHANED,
				RunID:    keyParts[3],
				NodeSpec: node.Spec,
			}
		}
	}

	return
}

func (r *Repository) getRun(ctx context.Context, id string, ops ...clientv3.OpOption) (*adagio.Run, error) {
	run := &adagio.Run{Id: id}

	resp, err := r.kv.Get(ctx, runKey(run), ops...)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) < 1 {
		return nil, errors.New("run not found")
	}

	// initially read out node configuration
	if err := unmarshalRun(resp.Kvs[0].Value, run); err != nil {
		return nil, err
	}

	// re-hydrate current node states
	if err := r.nodesForRun(ctx, run, ops...); err != nil {
		return nil, err
	}

	// check if all node states in order to derive run state
	var (
		runRunning   = false
		runCompleted = true
	)

	for _, node := range run.Nodes {
		runRunning = runRunning || (node.Status > adagio.Node_WAITING)
		runCompleted = runCompleted && (node.Status == adagio.Node_COMPLETED)
	}

	if runRunning {
		run.Status = adagio.Run_RUNNING
	}

	if runCompleted {
		run.Status = adagio.Run_COMPLETED
	}

	return run, nil
}

func (r *Repository) nodesForRun(ctx context.Context, run *adagio.Run, ops ...clientv3.OpOption) error {
	// ensure all calls use with prefix
	ops = append(ops, clientv3.WithPrefix())

	var (
		prefix    = allNodesKey(run)
		nodes     = map[string]*adagio.Node{}
		resp, err = r.kv.Get(ctx, prefix, ops...)
	)

	if err != nil {
		return err
	}

	// given no options are supplied then enforce all nodes header revision
	// is used for subsequent calls
	// < 2 because we added with prefix
	if len(ops) < 2 {
		ops = append(ops, clientv3.WithRev(resp.Header.Revision))
	}

	// create mapping for existing nodes in-order
	// to replace with deserialized ones
	for _, node := range run.Nodes {
		nodes[node.Spec.Name] = node
	}

	for _, kv := range resp.Kvs {
		node := &adagio.Node{}
		if err := json.Unmarshal(kv.Value, node); err != nil {
			return err
		}

		// check status key exists at store revision for each node
		resp, err := r.kv.Get(ctx, nodeInStateKey(run.Id, statusToString(node.Status), node.Spec.Name), ops...)
		if err != nil {
			return err
		}

		if len(resp.Kvs) < 1 {
			// node not found with expected status suggests node
			// has been orphaned so we mark it with none status
			node.Status = adagio.Node_NONE
		}

		if dst, ok := nodes[node.Spec.Name]; ok {
			// swap existing node with new de-serialized version
			*dst = *node
		}
	}

	for _, node := range run.Nodes {
		if err = r.setInputs(ctx, run, node); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) setInputs(ctx context.Context, run *adagio.Run, node *adagio.Node) error {
	// for each incoming node fetch their outputs
	incoming, err := adagio.GraphFrom(run).Incoming(node)
	if err != nil {
		return err
	}

	for ini := range incoming {
		in := ini.(*adagio.Node)

		adagio.VisitLatestAttempt(in, func(result *adagio.Node_Result) {
			if result.Conclusion != adagio.Node_Result_SUCCESS {
				return
			}

			if node.Inputs == nil {
				node.Inputs = map[string][]byte{}
			}

			node.Inputs[in.Spec.Name] = result.Output
		})
	}

	return nil
}

// UnsubscribeAll unsubscribes the provided agent and channel as a listener
func (r *Repository) UnsubscribeAll(ctx context.Context, a *adagio.Agent, ch chan<- *adagio.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// delete agent key before the
	_, err := r.kv.Delete(ctx, agentKey(a))
	if err != nil {
		return err
	}

	// release agent key lease when possible
	r.cancelLease(a.Id)

	if signal, ok := r.subscriptions[ch]; ok {
		signal <- struct{}{}

		delete(r.subscriptions, ch)
	}

	return nil
}

func agentKey(agent *adagio.Agent) string {
	return agentsPrefix + agent.Id
}

func runKey(run *adagio.Run) string {
	return runsPrefix + run.Id
}

func allNodesKey(run *adagio.Run) string {
	return fmt.Sprintf("%s%s/node/", nodesPrefix, run.Id)
}

func nodeKey(runID, name string) string {
	return fmt.Sprintf("%s%s/node/%s", nodesPrefix, runID, name)
}

func nodesInStateKey(status adagio.Node_Status) string {
	return statesPrefix + statusToString(status)
}

func nodeInStateKey(runID, state, name string) string {
	return fmt.Sprintf("%s%s/run/%s/node/%s", statesPrefix, state, runID, name)
}

type run struct {
	CreatedAt time.Time           `json:"created_at"`
	Specs     []*adagio.Node_Spec `json:"specs"`
	Edges     []*adagio.Edge      `json:"edges"`
}

func unmarshalRun(data []byte, dst *adagio.Run) error {
	var run run
	if err := json.Unmarshal(data, &run); err != nil {
		return nil
	}

	dst.CreatedAt = run.CreatedAt.Format(time.RFC3339Nano)
	dst.Edges = run.Edges

	// create an initial specification with zeroed node state
	// which will be replaced when nodes fetched and de-serialized
	for _, spec := range run.Specs {
		dst.Nodes = append(dst.Nodes, &adagio.Node{Spec: spec})
	}

	return nil
}

func marshalRun(createdAt string, spec []*adagio.Node_Spec, edges []*adagio.Edge) ([]byte, error) {
	var (
		createdAtT, err = time.Parse(time.RFC3339Nano, createdAt)
		run             = run{createdAtT, spec, edges}
	)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&run)
}

func statusToString(status adagio.Node_Status) string {
	return strings.ToLower(status.String())
}

func stringToStatus(status string) adagio.Node_Status {
	return adagio.Node_Status(adagio.Node_Status_value[strings.ToUpper(status)])
}

type types []adagio.Event_Type

func (t types) contains(typ adagio.Event_Type) bool {
	for _, needle := range t {
		if needle == typ {
			return true
		}
	}

	return false
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

type oneReader struct{}

func (oneReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 255
	}
	return len(p), nil
}
