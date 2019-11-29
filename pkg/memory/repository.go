package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/agent"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/service/controlplane"
)

var (
	// compile time check to ensure Repository is a agent.Repository
	_ agent.Repository        = (*Repository)(nil)
	_ controlplane.Repository = (*Repository)(nil)
)

type (
	listenerSet map[adagio.Event_Type][]chan<- *adagio.Event

	runState struct {
		run    *adagio.Run
		lookup map[string]*adagio.Node
		graph  *graph.Graph
	}
)

// Repository is an in-memory implementation of the adagio Repository interfaces
// It adheres to the repository test harness
type Repository struct {
	agents map[string]*adagio.Agent
	runs   map[string]runState
	claims map[string]struct {
		run  *adagio.Run
		node *adagio.Node
	}

	listeners listenerSet
	mu        sync.Mutex

	now func() time.Time
}

// New constructs and configures a new in memory repository
func New() *Repository {
	return &Repository{
		agents: map[string]*adagio.Agent{},
		runs:   map[string]runState{},
		claims: map[string]struct {
			run  *adagio.Run
			node *adagio.Node
		}{},
		listeners: listenerSet{},
	}
}

// Stats returns counts of runs and nodes in their respective states
func (r *Repository) Stats(context.Context) (*adagio.Stats, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodeCounts := &adagio.Stats_NodeCounts{}

	for _, runState := range r.runs {
		for _, node := range runState.lookup {
			switch node.Status {
			case adagio.Node_WAITING:
				nodeCounts.WaitingCount++
			case adagio.Node_READY:
				nodeCounts.ReadyCount++
			case adagio.Node_RUNNING:
				nodeCounts.RunningCount++
			case adagio.Node_COMPLETED:
				nodeCounts.CompletedCount++
			}
		}
	}

	return &adagio.Stats{
		RunCount:   int64(len(r.runs)),
		NodeCounts: nodeCounts,
	}, nil
}

// StartRun instantiates a run from a provided graph specification
func (r *Repository) StartRun(_ context.Context, spec *adagio.GraphSpec) (run *adagio.Run, err error) {
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
			r.notifyReady(run, node)
		}
	}

	return
}

// InspectRun returns a run for the provided run ID
func (r *Repository) InspectRun(_ context.Context, id string) (*adagio.Run, error) {
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

// ListAgents returns a set of subscribed agents
func (r *Repository) ListAgents(_ context.Context) (agents []*adagio.Agent, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Id < agents[j].Id
	})

	return
}

// ListRuns returns a list of runs in descending order based on a set of provided predicates
func (r *Repository) ListRuns(_ context.Context, req controlplane.ListRequest) (runs []*adagio.Run, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, state := range r.runs {
		runs = append(runs, state.run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Id > runs[j].Id
	})

	if req.Start != nil || req.Finish != nil {
		var (
			min            int
			max            = len(runs)
			minSet, maxSet bool
			start, _       = time.Parse(time.RFC3339, runs[min].CreatedAt)
		)

		finish, terr := time.Parse(time.RFC3339, runs[max-1].CreatedAt)
		if terr != nil {
			finish = time.Unix(0, math.MaxInt64)
		}

		if req.Start != nil {
			start = *req.Start
		}

		if req.Finish != nil {
			finish = *req.Finish
		}

		if start.Before(finish) {
			// start must be > finish as time descending
			runs = nil
			return
		}

		for i, run := range runs {
			createdAt, err := time.Parse(time.RFC3339Nano, run.CreatedAt)
			if err != nil {
				continue
			}

			if !minSet && (createdAt.Before(start) || createdAt == start) {
				minSet = true
				min = i
			}

			if !maxSet && (createdAt.Before(finish) || createdAt == finish) {
				maxSet = true
				max = i + 1
			}
		}

		runs = runs[min:max]
	}

	if limit := req.Limit; limit != nil {
		runs = runs[:int(*limit)]
	}

	return
}

// ClaimNode attempts to make a claim for a node
func (r *Repository) ClaimNode(_ context.Context, runID, name string, claim *adagio.Claim) (*adagio.Node, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, err := r.state(runID)
	if err != nil {
		return nil, false, err
	}

	node, ok := state.lookup[name]
	if !ok {
		return nil, false, fmt.Errorf("in-memory repository: node %q: %w", name, adagio.ErrMissingNode)
	}

	if node.Status == adagio.Node_WAITING {
		return nil, false, fmt.Errorf("in-memory repository: node %q: %w", name, adagio.ErrNodeNotReady)
	}

	// node already claimed
	if node.Status > adagio.Node_READY {
		return nil, false, nil
	}

	// update node state to running
	node.Status = adagio.Node_RUNNING
	node.StartedAt = r.now().Format(time.RFC3339Nano)
	node.Claim = claim

	r.claims[claim.Id] = struct {
		run  *adagio.Run
		node *adagio.Node
	}{state.run, node}

	return node, true, nil
}

func (r *Repository) notifyReady(run *adagio.Run, node *adagio.Node) {
	for _, ch := range r.listeners[adagio.Event_NODE_READY] {
		select {
		case ch <- &adagio.Event{RunID: run.Id, NodeSpec: node.Spec, Type: adagio.Event_NODE_READY}:
			// attempt to send
		default:
		}
	}
}

// FinishNode reports the result of a node run and readies any eligible outgoing nodes
func (r *Repository) FinishNode(_ context.Context, runID, name string, result *adagio.Node_Result, claim *adagio.Claim) error {
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
	node.FinishedAt = r.now().Format(time.RFC3339Nano)
	node.Attempts = append(node.Attempts, result)

	outgoing, err := state.graph.Outgoing(node)
	if err != nil {
		return fmt.Errorf("finishing node %q: %w", node, err)
	}

	delete(r.claims, claim.Id)

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
			return fmt.Errorf("finishing node %q: %w", node, err)
		}

		// given all the incoming nodes into "out" are now completed
		// then the waiting out node can be progressed to ready
		ready := true
		for in := range incoming {
			ready = ready && in.(*adagio.Node).Status == adagio.Node_COMPLETED
		}

		if ready {
			out.Status = adagio.Node_READY

			r.notifyReady(state.run, out)
		}
	}

	return nil
}

func (r *Repository) handleFailure(state runState, node *adagio.Node, src map[graph.Node]struct{}, result *adagio.Node_Result) error {
	if adagio.CanRetry(node) {
		// put node back into the ready state to be attempted again
		node.Status = adagio.Node_READY
		node.FinishedAt = ""

		r.notifyReady(state.run, node)

		return nil
	}

	// no attempts remaining so progress outgoing nodes into
	// completed but inconcluded state
	for outi := range src {
		out := outi.(*adagio.Node)

		out.Status = adagio.Node_COMPLETED
		out.StartedAt = r.now().Format(time.RFC3339Nano)
		out.FinishedAt = r.now().Format(time.RFC3339Nano)

		outgoing, err := state.graph.Outgoing(out)
		if err != nil {
			return fmt.Errorf("finishing node %q: %w", node, err)
		}

		// descend into child nodes
		if err := r.handleFailure(state, node, outgoing, result); err != nil {
			return err
		}
	}

	return nil
}

// Subscribe registers the provided channel to listen for the defined event types
func (r *Repository) Subscribe(_ context.Context, agent *adagio.Agent, events chan<- *adagio.Event, types ...adagio.Event_Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents[agent.Id] = agent

	for _, typ := range types {
		if typ == adagio.Event_NODE_READY {
			for _, state := range r.runs {
				for _, node := range state.lookup {
					if node.Status == adagio.Node_READY {
						events <- &adagio.Event{
							RunID:    state.run.Id,
							NodeSpec: node.Spec,
							Type:     adagio.Event_NODE_READY,
						}
					}
				}
			}
		}
		r.listeners[typ] = append(r.listeners[typ], events)
	}

	return nil
}

// UnsubscribeAll unsubscribes the channel for all event types
func (r *Repository) UnsubscribeAll(_ context.Context, agent *adagio.Agent, events chan<- *adagio.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.agents, agent.Id)

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
		return runState{}, fmt.Errorf("in-memory repository: run %q: %w", runID, adagio.ErrRunDoesNotExist)
	}

	return state, nil
}

func node(state runState, name string) (*adagio.Node, error) {
	node, ok := state.lookup[name]
	if !ok {
		return nil, fmt.Errorf("in-memory repository: node %q: %w", name, adagio.ErrMissingNode)
	}

	return node, nil
}
