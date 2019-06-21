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
	"github.com/georgemac/adagio/pkg/worker"
	"go.etcd.io/etcd/clientv3"
)

var (
	_ worker.Repository = (*Repository)(nil)
)

type Repository struct {
	kv      clientv3.KV
	watcher clientv3.Watcher

	mu            sync.Mutex
	subscriptions map[chan<- *adagio.Event]chan struct{}

	ns  namespace
	now func() time.Time
}

func New(kv clientv3.KV, watcher clientv3.Watcher, opts ...Option) *Repository {
	r := &Repository{
		kv:            kv,
		watcher:       watcher,
		mu:            sync.Mutex{},
		subscriptions: map[chan<- *adagio.Event]chan struct{}{},
		now:           func() time.Time { return time.Now().UTC() },
		ns:            namespace("v0/"),
	}

	Options(opts).Apply(r)

	return r
}

func (r *Repository) StartRun(spec *adagio.GraphSpec) (run *adagio.Run, err error) {
	run, err = adagio.NewRun(spec)
	if err != nil {
		return
	}

	data, err := marshalRun(run.CreatedAt, run.Edges)
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

		state, err := stateToString(node.Status)
		if err != nil {
			return nil, err
		}

		var (
			key = r.ns.nodeInStateKey(run.Id, state, node.Spec.Name)
			put = clientv3.OpPut(key, string(nodeData))
		)

		ops = append(ops, put)
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
		parts := strings.SplitN(r.ns.stripBytes(kv.Key), "/", 2)
		if len(parts) < 2 {
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

func (r *Repository) ClaimNode(runID, name string) (*adagio.Node, bool, error) {
	ctxt := context.Background()
	run, err := r.getRun(ctxt, runID)
	if err != nil {
		return nil, false, err
	}

	node, err := run.GetNodeByName(name)
	if err != nil {
		return nil, false, err
	}

	if node.Status != adagio.Node_READY {
		return nil, false, adagio.ErrNodeNotReady
	}

	cmps, ops, err := r.transition(ctxt, runID, node, adagio.Node_RUNNING)
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

	// for each incoming node fetch their outputs
	incoming, err := adagio.GraphFrom(run).Incoming(node)
	if err != nil {
		return nil, false, err
	}

	for ini := range incoming {
		in := ini.(*adagio.Node)

		resp, err := r.kv.Get(ctxt, r.ns.output(runID, in.Spec.Name))
		if err != nil {
			return nil, false, err
		}

		if len(resp.Kvs) > 0 {
			if node.Inputs == nil {
				node.Inputs = map[string][]byte{}
			}

			node.Inputs[in.Spec.Name] = resp.Kvs[0].Value
		}
	}

	return node, resp.Succeeded, nil
}

func (r *Repository) transition(ctxt context.Context, runID string, node *adagio.Node, toStatus adagio.Node_Status) ([]clientv3.Cmp, []clientv3.Op, error) {
	from, err := stateToString(node.Status)
	if err != nil {
		return nil, nil, err
	}

	to, err := stateToString(toStatus)
	if err != nil {
		return nil, nil, err
	}

	var (
		fromKey = r.ns.nodeInStateKey(runID, from, node.Spec.Name)
		toKey   = r.ns.nodeInStateKey(runID, to, node.Spec.Name)
	)

	node.Status = toStatus

	switch toStatus {
	case adagio.Node_RUNNING:
		node.StartedAt = r.now().Format(time.RFC3339)
	case adagio.Node_COMPLETED:
		node.FinishedAt = r.now().Format(time.RFC3339)
	}

	data, err := json.Marshal(node)
	if err != nil {
		return nil, nil, err
	}

	return []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(fromKey), ">", 0),
			clientv3.Compare(clientv3.Version(toKey), "=", 0),
		}, []clientv3.Op{
			clientv3.OpPut(toKey, string(data)),
			clientv3.OpDelete(fromKey),
		}, nil
}

func (r *Repository) FinishNode(runID, name string, result *adagio.Result) error {
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

	graph := adagio.GraphFrom(run)

	cmps, ops, err := r.transition(ctxt, runID, node, adagio.Node_COMPLETED)
	if err != nil {
		return err
	}

	ops = append(ops, clientv3.OpPut(r.ns.output(run.Id, node.Spec.Name), string(result.Output)))

	outgoing, err := graph.Outgoing(node)
	if err != nil {
		return err
	}

	for o := range outgoing {
		out := o.(*adagio.Node)

		isReady := true

		incoming, err := graph.Incoming(out)
		if err != nil {
			return err
		}

		for v := range incoming {
			in := v.(*adagio.Node)

			isReady = isReady && in.Status == adagio.Node_COMPLETED

			if in == node {
				continue
			}

			state, err := stateToString(in.Status)
			if err != nil {
				return err
			}

			currentKey := r.ns.nodeInStateKey(runID, state, in.Spec.Name)
			cmps = append(cmps, clientv3.Compare(clientv3.Version(currentKey), ">", 0))
		}

		// if all nodes are now completed
		// then the outgoing target is ready
		if isReady {
			outCmps, outOps, err := r.transition(ctxt, runID, out, adagio.Node_READY)
			if err != nil {
				return err
			}

			cmps = append(cmps, outCmps...)
			ops = append(ops, outOps...)
		}
	}

	resp, err := r.kv.Txn(ctxt).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return r.FinishNode(runID, name, result)
	}

	return nil
}

func (r *Repository) Subscribe(events chan<- *adagio.Event, s ...adagio.Node_Status) error {
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

			if len(resp.Events) > 0 {
				for _, ev := range resp.Events {
					if ev.IsCreate() {
						keyParts := strings.Split(r.ns.stripBytes(ev.Kv.Key), "/")
						if len(keyParts) < 6 {
							continue
						}

						run, err := r.getRun(context.Background(), keyParts[3])
						if err != nil {
							log.Println(err)
							continue
						}

						node, err := run.GetNodeByName(keyParts[5])
						if err != nil {
							log.Println(err)
							continue
						}

						state, err := stateFromString(keyParts[1])
						if err != nil {
							log.Println(err)
							continue
						}

						if states(s).contains(state) {
							events <- &adagio.Event{
								Type:     adagio.Event_STATE_TRANSITION,
								RunID:    keyParts[3],
								NodeSpec: node.Spec,
							}
						}
					}
				}
			}
		}
	}()

	return nil
}

func (r *Repository) getRun(ctxt context.Context, id string) (*adagio.Run, error) {
	run := adagio.Run{
		Id: id,
	}

	resp, err := r.kv.Get(ctxt, r.ns.runKey(&adagio.Run{Id: id}))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) < 1 {
		return nil, errors.New("run not found")
	}

	if err := unmarshalRun(resp.Kvs[0].Value, &run); err != nil {
		return nil, err
	}

	run.Nodes, err = r.nodesForRun(ctxt, run.Id)
	if err != nil {
		return nil, err
	}

	return &run, nil
}

func (r *Repository) getNode(ctxt context.Context, key string) (*adagio.Node, error) {
	node := &adagio.Node{}

	resp, err := r.kv.Get(ctxt, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) < 1 {
		return nil, adagio.ErrMissingNode
	}

	return node, json.Unmarshal(resp.Kvs[0].Value, node)
}

func (r *Repository) nodesForRun(ctxt context.Context, id string) (nodes []*adagio.Node, err error) {
	for _, state := range []string{"waiting", "ready", "running", "completed"} {
		var (
			prefix    = r.ns.nodesInStateKey(id, state)
			resp, err = r.kv.Get(ctxt, prefix, clientv3.WithPrefix())
		)

		if err != nil {
			return nil, err
		}

		for _, kv := range resp.Kvs {
			node := &adagio.Node{}
			if err := json.Unmarshal(kv.Value, node); err != nil {
				return nil, err
			}

			nodes = append(nodes, node)
		}
	}

	return
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

func (n namespace) nodesInStateKey(runID, state string) string {
	return fmt.Sprintf("%sstates/%s/run/%s", n, state, runID)
}

func (n namespace) nodeInStateKey(runID, state, name string) string {
	return fmt.Sprintf("%sstates/%s/run/%s/node/%s", n, state, runID, name)
}

func (n namespace) output(runID, node string) string {
	return fmt.Sprintf("%sruns/%s/node/%s/output", n, runID, node)
}

func (n namespace) stripBytes(key []byte) string {
	return strings.TrimPrefix(string(key), string(n))
}

type run struct {
	CreatedAt time.Time      `json:"created_at"`
	Edges     []*adagio.Edge `json:"edges"`
}

func unmarshalRun(data []byte, dst *adagio.Run) error {
	var run run
	if err := json.Unmarshal(data, &run); err != nil {
		return nil
	}

	dst.CreatedAt = run.CreatedAt.Format(time.RFC3339)
	dst.Edges = run.Edges

	return nil
}

func marshalRun(createdAt string, edges []*adagio.Edge) ([]byte, error) {
	var (
		createdAtT, err = time.Parse(time.RFC3339, createdAt)
		run             = run{createdAtT, edges}
	)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&run)
}

type states []adagio.Node_Status

func (s states) contains(state adagio.Node_Status) bool {
	for _, needle := range s {
		if state == needle {
			return true
		}
	}

	return false
}

func stateFromString(state string) (adagio.Node_Status, error) {
	switch state {
	case "waiting":
		return adagio.Node_WAITING, nil
	case "ready":
		return adagio.Node_READY, nil
	case "running":
		return adagio.Node_RUNNING, nil
	case "completed":
		return adagio.Node_COMPLETED, nil
	default:
		return adagio.Node_WAITING, errors.New("state not recognized")
	}
}

func stateToString(state adagio.Node_Status) (string, error) {
	switch state {
	case adagio.Node_WAITING:
		return "waiting", nil
	case adagio.Node_READY:
		return "ready", nil
	case adagio.Node_RUNNING:
		return "running", nil
	case adagio.Node_COMPLETED:
		return "completed", nil
	default:
		return "", errors.New("state not recognized")
	}
}
