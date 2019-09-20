package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/worker"
	"go.etcd.io/etcd/clientv3"
)

var (
	_ worker.Repository = (*Repository)(nil)
)

type Repository struct {
	kv      clientv3.KV
	watcher clientv3.Watcher
	leaser  clientv3.Lease

	mu            sync.Mutex
	subscriptions map[chan<- *adagio.Event]chan struct{}

	ns  namespace
	now func() time.Time

	ttl    time.Duration
	leases map[string]func()
}

func New(kv clientv3.KV, watcher clientv3.Watcher, leaser clientv3.Lease, opts ...Option) *Repository {
	r := &Repository{
		kv:            kv,
		watcher:       watcher,
		leaser:        leaser,
		mu:            sync.Mutex{},
		subscriptions: map[chan<- *adagio.Event]chan struct{}{},
		now:           func() time.Time { return time.Now().UTC() },
		ns:            namespace("v0/"),
		ttl:           10 * time.Second,
		leases:        map[string]func(){},
	}

	Options(opts).Apply(r)

	return r
}

func (r *Repository) StartRun(spec *adagio.GraphSpec) (run *adagio.Run, err error) {
	run, err = adagio.NewRun(spec)
	if err != nil {
		return
	}

	data, err := marshalRun(run.CreatedAt, spec.Nodes, run.Edges)
	if err != nil {
		return nil, err
	}

	var (
		runKey = r.ns.runKey(run)
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
			key      = r.ns.nodeKey(run.Id, node.Spec.Name)
			stateKey = r.ns.nodeInStateKey(run.Id, statusToString(node.Status), node.Spec.Name)
			put      = clientv3.OpPut(key, string(nodeData))
			putState = clientv3.OpPut(stateKey, "")
		)

		ops = append(ops, put, putState)
	}

	resp, err := r.kv.Txn(context.Background()).
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

func (r *Repository) InspectRun(id string) (*adagio.Run, error) {
	return r.getRun(context.Background(), id)
}

func (r *Repository) ListRuns() (runs []*adagio.Run, err error) {
	ctxt := context.Background()
	resp, err := r.kv.Get(ctxt, r.ns.runs(), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		// strip namespace from key
		parts := strings.Split(r.ns.stripBytes(kv.Key), "/")

		// ignore non-run keys
		if len(parts) != 2 {
			continue
		}

		run, err := r.getRun(ctxt, parts[1])
		if err != nil {
			return nil, err
		}

		runs = append(runs, run)
	}

	return
}

func (r *Repository) ClaimNode(runID, name string, claim *adagio.Claim) (node *adagio.Node, claimed bool, err error) {
	ctxt := context.Background()
	run, err := r.getRun(ctxt, runID)
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

	cmps, ops, err := r.transition(ctxt, runID, node, adagio.Node_RUNNING, nil, nil, clientv3.WithLease(leaseID))
	if err != nil {
		return nil, false, err
	}

	resp, err := r.kv.Txn(ctxt).
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

func (r *Repository) complete(ctxt context.Context, runID string, node *adagio.Node, cmps []clientv3.Cmp, ops []clientv3.Op) ([]clientv3.Cmp, []clientv3.Op, error) {
	// given node has not already been completed
	if node.Status != adagio.Node_COMPLETED {
		var err error
		cmps, ops, err = r.transition(ctxt, runID, node, adagio.Node_COMPLETED, cmps, ops)
		if err != nil {
			return nil, nil, err
		}
	}

	return cmps, ops, nil
}

func (r *Repository) transition(ctxt context.Context, runID string, node *adagio.Node, toStatus adagio.Node_Status, cmps []clientv3.Cmp, ops []clientv3.Op, putOpts ...clientv3.OpOption) ([]clientv3.Cmp, []clientv3.Op, error) {
	var (
		nodeKey = r.ns.nodeKey(runID, node.Spec.Name)
		fromKey = r.ns.nodeInStateKey(runID, statusToString(node.Status), node.Spec.Name)
		toKey   = r.ns.nodeInStateKey(runID, statusToString(toStatus), node.Spec.Name)
		from    = node.Status
	)

	node.Status = toStatus

	switch toStatus {
	case adagio.Node_RUNNING:
		node.StartedAt = r.now().Format(time.RFC3339)

	case adagio.Node_COMPLETED:
		now := r.now()
		if node.StartedAt == "" {
			node.StartedAt = now.Format(time.RFC3339)
		}

		node.FinishedAt = now.Format(time.RFC3339)
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

	// the following code handles orphaned nodes where
	// the node is now not in any state
	statusDoesNotExist := func(status adagio.Node_Status) clientv3.Cmp {
		statusKey := r.ns.nodeInStateKey(runID, statusToString(status), node.Spec.Name)
		return clientv3.Compare(clientv3.Version(statusKey), "=", 0)
	}

	return append(cmps,
			// ensure node exists
			clientv3.Compare(clientv3.Version(nodeKey), ">", 0),
			// ensure node does not exist in any state
			statusDoesNotExist(adagio.Node_WAITING),
			statusDoesNotExist(adagio.Node_READY),
			statusDoesNotExist(adagio.Node_RUNNING),
			statusDoesNotExist(adagio.Node_COMPLETED)),
		append(ops,
			clientv3.OpPut(nodeKey, string(data)),
			clientv3.OpPut(toKey, "", putOpts...),
		), nil
}

func (r *Repository) lease(claimID string) (clientv3.LeaseID, error) {
	// store lease keep-alive cancel func
	ctxt, cancel := context.WithCancel(context.Background())

	// grant lease in seconds
	leaseResp, err := r.leaser.Grant(ctxt, int64(r.ttl/time.Second))
	if err != nil {
		return 0, err
	}

	r.mu.Lock()
	r.leases[claimID] = func() {
		// cancel keep-alive
		cancel()

		// revoke lease
		if _, err := r.leaser.Revoke(context.Background(), leaseResp.ID); err != nil {
			log.Println(claimID, err)
			return
		}
	}
	r.mu.Unlock()

	// keep lease alive
	go func() {
		defer r.cancelLease(claimID)

		resps, err := r.leaser.KeepAlive(ctxt, leaseResp.ID)
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
	r.mu.Lock()
	if cancel, ok := r.leases[claimID]; ok {
		cancel()
		delete(r.leases, claimID)
	}
	r.mu.Unlock()
}

func (r *Repository) FinishNode(runID, name string, result *adagio.Node_Result, claim *adagio.Claim) error {
	ctxt := context.Background()
	run, err := r.getRun(ctxt, runID)
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
		cmps, ops, err = r.handleSuccess(ctxt, run, node, result)
		if err != nil {
			return err
		}
	} else {
		cmps, ops, err = r.handleFailure(ctxt, run, node, result)
		if err != nil {
			return err
		}
	}

	resp, err := r.kv.Txn(ctxt).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		r.cancelLease(claim.Id)

		return err
	}

	if !resp.Succeeded {
		return r.FinishNode(run.Id, node.Spec.Name, result, claim)
	}

	r.cancelLease(claim.Id)

	return nil
}

func (r *Repository) handleSuccess(ctxt context.Context, run *adagio.Run, node *adagio.Node, result *adagio.Node_Result) ([]clientv3.Cmp, []clientv3.Op, error) {
	cmps, ops, err := r.complete(ctxt, run.Id, node, nil, nil)
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
			// outgoing node has already progressed from waiting state
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
				continue
			}

			currentKey := r.ns.nodeInStateKey(run.Id, statusToString(in.Status), in.Spec.Name)
			cmps = append(cmps, clientv3.Compare(clientv3.Version(currentKey), ">", 0))
		}

		// if all nodes are now completed
		// then the outgoing target is ready
		if isReady {
			cmps, ops, err = r.transition(ctxt, run.Id, out, adagio.Node_READY, cmps, ops)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return cmps, ops, nil
}

func (r *Repository) handleFailure(ctxt context.Context, run *adagio.Run, node *adagio.Node, result *adagio.Node_Result) ([]clientv3.Cmp, []clientv3.Op, error) {
	if adagio.CanRetry(node) {
		// put node back into the ready state to be attempted again
		return r.transition(ctxt, run.Id, node, adagio.Node_READY, nil, nil)
	}

	cmps, ops, err := r.complete(ctxt, run.Id, node, nil, nil)
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
		cmps, ops, err = r.complete(ctxt, run.Id, node, cmps, ops)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return cmps, ops, nil
}

func (r *Repository) Subscribe(events chan<- *adagio.Event, typ ...adagio.Event_Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan struct{})

	r.subscriptions[events] = ch

	go func() {
		watch := r.watcher.Watch(context.Background(), r.ns.states(), clientv3.WithPrefix())
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
				keyParts := strings.Split(r.ns.stripBytes(ev.Kv.Key), "/")
				if len(keyParts) < 6 || keyParts[0] != "states" {
					// we're only interested in keys that represent states
					continue
				}

				status := stringToStatus(keyParts[1])
				switch status {
				// we're only interested in actions on ready and running keys
				case adagio.Node_READY, adagio.Node_RUNNING:
				default:
					continue
				}

				run, err := r.getRun(context.Background(), keyParts[3], clientv3.WithRev(resp.Header.Revision))
				if err != nil {
					log.Println(keyParts[3], err)
					continue
				}

				node, err := run.GetNodeByName(keyParts[5])
				if err != nil {
					log.Println(err)
					continue
				}

				// if a ready status key has been created and the subscription contains a
				// node ready type then send a node ready event
				if ev.IsCreate() && status == adagio.Node_READY && types(typ).contains(adagio.Event_NODE_READY) {
					events <- &adagio.Event{
						Type:     adagio.Event_NODE_READY,
						RunID:    keyParts[3],
						NodeSpec: node.Spec,
					}

					continue
				}

				// todo(george): import issues with etcd returning old urls from types
				// defined in new url scheme makes importing other etcd v3 libs to break
				// until then I am just going to compare to the literal 1 value which means
				// DELETE operation
				// given the deletion of a running key where no other state key for the
				// node exists (this is where GetNodeByName returns a node with a NONE status)
				if ev.Type == 1 && status == adagio.Node_RUNNING && node.Status == adagio.Node_NONE {
					// only send if the subscription contains the node orphaned
					// event type
					if types(typ).contains(adagio.Event_NODE_ORPHANED) {
						events <- &adagio.Event{
							Type:     adagio.Event_NODE_ORPHANED,
							RunID:    keyParts[3],
							NodeSpec: node.Spec,
						}
					}
				}
			}
		}
	}()

	return nil
}

func (r *Repository) getRun(ctxt context.Context, id string, ops ...clientv3.OpOption) (*adagio.Run, error) {
	run := &adagio.Run{
		Id: id,
	}

	resp, err := r.kv.Get(ctxt, r.ns.runKey(run), ops...)
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
	if err := r.nodesForRun(ctxt, run, ops...); err != nil {
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

func (r *Repository) nodesForRun(ctxt context.Context, run *adagio.Run, ops ...clientv3.OpOption) error {
	// ensure all calls use with prefix
	ops = append(ops, clientv3.WithPrefix())

	var (
		prefix    = r.ns.allNodesKey(run)
		nodes     = map[string]*adagio.Node{}
		resp, err = r.kv.Get(ctxt, prefix, ops...)
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
		resp, err := r.kv.Get(ctxt, r.ns.nodeInStateKey(run.Id, statusToString(node.Status), node.Spec.Name), ops...)
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
		if err = r.setInputs(ctxt, run, node); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) setInputs(ctxt context.Context, run *adagio.Run, node *adagio.Node) error {
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

func (r *Repository) UnsubscribeAll(ch chan<- *adagio.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if signal, ok := r.subscriptions[ch]; ok {
		signal <- struct{}{}

		delete(r.subscriptions, ch)
	}

	return nil
}

type namespace string

func (n namespace) runs() string {
	return string(n) + "runs/"
}

func (n namespace) states() string {
	return string(n) + "states/"
}

func (n namespace) runKey(run *adagio.Run) string {
	return fmt.Sprintf("%sruns/%s", n, run.Id)
}

func (n namespace) allNodesKey(run *adagio.Run) string {
	return fmt.Sprintf("%sruns/%s/node/", n, run.Id)
}

func (n namespace) nodeKey(runID, name string) string {
	return fmt.Sprintf("%sruns/%s/node/%s", n, runID, name)
}

func (n namespace) nodeInStateKey(runID, state, name string) string {
	return fmt.Sprintf("%sstates/%s/run/%s/node/%s", n, state, runID, name)
}

func (n namespace) stripBytes(key []byte) string {
	return strings.TrimPrefix(string(key), string(n))
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

	dst.CreatedAt = run.CreatedAt.Format(time.RFC3339)
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
		createdAtT, err = time.Parse(time.RFC3339, createdAt)
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
