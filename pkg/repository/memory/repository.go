package memory

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/repository"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
)

// compile time check to ensure Repository is a repository.Repository
var _ repository.Repository = (*Repository)(nil)

type (
	nodeSet     map[*adagio.Node]struct{}
	listenerSet map[adagio.NodeState][]chan<- repository.Event
)

type Repository struct {
	runs    map[string]*adagio.Run
	waiting map[*adagio.Node]nodeSet

	ready,
	running,
	done,
	dead nodeSet

	listeners listenerSet

	entropy io.Reader

	mu sync.Mutex
}

func NewRepository() *Repository {
	return &Repository{
		runs:      map[string]*adagio.Run{},
		waiting:   map[*adagio.Node]nodeSet{},
		ready:     nodeSet{},
		running:   nodeSet{},
		done:      nodeSet{},
		dead:      nodeSet{},
		listeners: listenerSet{},
		entropy:   ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}
}

func (r *Repository) StartRun(graph adagio.Graph) (run *adagio.Run, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	run = &adagio.Run{
		ID:    ulid.MustNew(ulid.Timestamp(now), r.entropy).String(),
		Graph: graph,
	}

	r.runs[run.ID] = run

	run.Graph.Walk(func(node *adagio.Node) {
		if ready, _ := run.Graph.IsRoot(node); ready {
			r.notifyListeners(run, node, adagio.NoneState, adagio.ReadyState)
			r.ready[node] = struct{}{}
			return
		}

		r.notifyListeners(run, node, adagio.NoneState, adagio.WaitingState)
		r.waiting[node], err = run.Graph.Incoming(node)
		if err != nil {
			return
		}
	})

	return
}

func (r *Repository) ListRuns() (runs []*adagio.Run, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, run := range r.runs {
		runs = append(runs, run)
	}

	return
}

func (r *Repository) ClaimNode(run *adagio.Run, node *adagio.Node) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.runMustExist(run); err != nil {
		return false, err
	}

	if err := r.mustNotBeWaiting(node); err != nil {
		return false, err
	}

	if _, ok := r.ready[node]; !ok {
		// node not ready
		return false, nil
	}

	delete(r.ready, node)

	r.running[node] = struct{}{}

	r.notifyListeners(run, node, adagio.ReadyState, adagio.RunningState)

	return true, nil
}

func (r *Repository) notifyListeners(run *adagio.Run, node *adagio.Node, from, to adagio.NodeState) {
	for _, ch := range r.listeners[to] {
		select {
		case ch <- repository.Event{run, node, from, to}:
			// attempt to send
		default:
		}
	}
}

func (r *Repository) FinishNode(run *adagio.Run, node *adagio.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.runMustExist(run); err != nil {
		return err
	}

	if _, ok := r.running[node]; !ok {
		return errors.New("cannot finish node in this state")
	}

	outgoing, err := run.Graph.Outgoing(node)
	if err != nil {
		return errors.Wrapf(err, "finishing node %q", node)
	}

	for out := range outgoing {
		delete(r.waiting[out], node)

		if len(r.waiting[out]) == 0 {
			// if it no longer is blocked by anything
			delete(r.waiting, out)
			r.ready[out] = struct{}{}

			r.notifyListeners(run, out, adagio.WaitingState, adagio.ReadyState)
		}
	}

	delete(r.running, node)

	r.done[node] = struct{}{}

	return nil
}

func (r *Repository) RecoverNode(*adagio.Run, *adagio.Node) (bool, error) {
	panic("not implemented")
}

func (r *Repository) BuryNode(*adagio.Run, *adagio.Node) error {
	panic("not implemented")
}

func (r *Repository) Subscribe(events chan<- repository.Event, states ...adagio.NodeState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, state := range states {
		r.listeners[state] = append(r.listeners[state], events)
	}

	return nil
}

func (r *Repository) runMustExist(run *adagio.Run) error {
	if _, ok := r.runs[run.ID]; !ok {
		return errors.Wrapf(repository.ErrRunDoesNotExist, "in-memory repository: run %q", run)
	}

	return nil
}

func (r *Repository) mustNotBeWaiting(node *adagio.Node) error {
	if _, ok := r.waiting[node]; ok {
		return errors.Wrapf(repository.ErrNodeNotReady, "in-memory repository: node %q", node)
	}

	return nil
}
