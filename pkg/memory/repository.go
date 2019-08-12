package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/pkg/errors"
)

var (
	// compile time check to ensure Repository is a worker.Repository
	_ worker.Repository = (*Repository)(nil)
)

type (
	listenerSet map[adagio.Event_Type][]chan<- *adagio.Event

	runState struct {
		run    *adagio.Run
		lookup map[string]*adagio.Node
		graph  *graph.Graph
	}
)

type Repository struct {
	runs map[string]runState

	listeners listenerSet
	mu        sync.Mutex

	now func() time.Time
}

func New() *Repository {
	return &Repository{
		runs:      map[string]runState{},
		listeners: listenerSet{},
	}
}

func (r *Repository) StartRun(spec *adagio.GraphSpec) (run *adagio.Run, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, err = adagio.NewRun(spec)
	if err != nil {
		return
	}

	state := runState{
		run:    run,
		lookup: map[string]*adagio.Node{},
		graph:  adagio.GraphFrom(run),
	}

	r.runs[run.Id] = state

	for _, node := range run.Nodes {
		state.lookup[node.Spec.Name] = node

		if node.Status == adagio.Node_READY {
			r.notifyTransition(run, node, adagio.Node_WAITING, adagio.Node_READY)
		}
	}

	return
}

func (r *Repository) InspectRun(id string) (*adagio.Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.state(id)
	if err != nil {
		return nil, err
	}

	run := state.run

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

func (r *Repository) ListRuns() (runs []*adagio.Run, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, state := range r.runs {
		runs = append(runs, state.run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Id < runs[j].Id
	})

	return
}

func (r *Repository) ClaimNode(runID, name string) (*adagio.Node, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.state(runID)
	if err != nil {
		return nil, false, err
	}

	node, ok := state.lookup[name]
	if !ok {
		return nil, false, errors.Wrapf(adagio.ErrMissingNode, "in-memory repository: node %q", name)
	}

	if node.Status == adagio.Node_WAITING {
		return nil, false, errors.Wrapf(adagio.ErrNodeNotReady, "in-memory repository: node %q", node)
	}

	// node already claimed
	if node.Status > adagio.Node_READY {
		return nil, false, nil
	}

	// update node state to running
	node.Status = adagio.Node_RUNNING
	node.StartedAt = r.now().Format(time.RFC3339)

	r.notifyTransition(state.run, node, adagio.Node_READY, adagio.Node_RUNNING)

	return node, true, nil
}

func (r *Repository) notifyTransition(run *adagio.Run, node *adagio.Node, from, to adagio.Node_Status) {
	for _, ch := range r.listeners[adagio.Event_STATE_TRANSITION] {
		select {
		case ch <- &adagio.Event{RunID: run.Id, NodeSpec: node.Spec, Type: adagio.Event_STATE_TRANSITION}:
			// attempt to send
		default:
		}
	}
}

func (r *Repository) FinishNode(runID, name string, result *adagio.Node_Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.state(runID)
	if err != nil {
		return err
	}

	node, err := node(state, name)
	if err != nil {
		return err
	}

	node.Status = adagio.Node_COMPLETED
	node.FinishedAt = r.now().Format(time.RFC3339)
	node.Attempts = append(node.Attempts, result)

	r.notifyTransition(state.run, node, adagio.Node_RUNNING, adagio.Node_COMPLETED)

	outgoing, err := state.graph.Outgoing(node)
	if err != nil {
		return errors.Wrapf(err, "finishing node %q", node)
	}

	if result.Conclusion == adagio.Node_Result_SUCCESS {
		return r.handleSuccess(state, node, outgoing, result)
	}

	return r.handleFailure(state, node, outgoing, result)
}

func (r *Repository) handleSuccess(state runState, node *adagio.Node, outgoing map[graph.Node]struct{}, result *adagio.Node_Result) error {
	for outi := range outgoing {
		out := outi.(*adagio.Node)

		// propagate outputs to inputs of next node
		if out.Inputs == nil {
			out.Inputs = map[string][]byte{}
		}

		out.Inputs[node.Spec.Name] = result.Output

		if out.Status > adagio.Node_WAITING {
			// do not bother to manipulate outgoing nodes which are not waiting
			continue
		}

		incoming, err := state.graph.Incoming(out)
		if err != nil {
			return errors.Wrapf(err, "finishing node %q", node)
		}

		// given all the incoming nodes into "out" are now completed
		// then the waiting out node can be progressed to ready
		ready := true
		for in := range incoming {
			ready = ready && in.(*adagio.Node).Status == adagio.Node_COMPLETED
		}

		if ready {
			out.Status = adagio.Node_READY

			r.notifyTransition(state.run, out, adagio.Node_WAITING, adagio.Node_READY)
		}
	}

	return nil
}

func (r *Repository) handleFailure(state runState, node *adagio.Node, src map[graph.Node]struct{}, result *adagio.Node_Result) error {
	if adagio.CanRetry(node) {
		// put node back into the ready state to be attempted again
		node.Status = adagio.Node_READY
		node.FinishedAt = ""

		r.notifyTransition(state.run, node, adagio.Node_RUNNING, adagio.Node_READY)

		return nil
	}

	// no attempts remaining so progress outgoing nodes into
	// completed but inconcluded state
	for outi := range src {
		out := outi.(*adagio.Node)

		out.Status = adagio.Node_COMPLETED
		out.StartedAt = r.now().Format(time.RFC3339)
		out.FinishedAt = r.now().Format(time.RFC3339)

		r.notifyTransition(state.run, out, out.Status, adagio.Node_COMPLETED)

		outgoing, err := state.graph.Outgoing(out)
		if err != nil {
			return errors.Wrapf(err, "finishing node %q", node)
		}

		// descend into child nodes
		if err := r.handleFailure(state, node, outgoing, result); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) Subscribe(events chan<- *adagio.Event, types ...adagio.Event_Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, typ := range types {
		r.listeners[typ] = append(r.listeners[typ], events)
	}

	return nil
}

func (r *Repository) UnsubscribeAll(events chan<- *adagio.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for event, chans := range r.listeners {
		for i, ch := range chans {
			if ch == events {
				// remove channel from listening map
				r.listeners[event] = append(chans[0:i], chans[i+1:]...)
			}
		}
	}

	return nil
}

func (r *Repository) state(runID string) (runState, error) {
	state, ok := r.runs[runID]
	if !ok {
		return runState{}, errors.Wrapf(adagio.ErrRunDoesNotExist, "in-memory repository: run %q", runID)
	}

	return state, nil
}

func node(state runState, name string) (*adagio.Node, error) {
	node, ok := state.lookup[name]
	if !ok {
		return nil, errors.Wrapf(adagio.ErrMissingNode, "in-memory repository: node %q", name)
	}

	return node, nil
}
