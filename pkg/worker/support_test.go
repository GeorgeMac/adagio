package worker

import (
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

type repository struct {
	mu sync.Mutex

	// expectation of the number of subscriptions
	subscriptionCount sync.WaitGroup
	// return values
	nodes map[string]*adagio.Node
	// calls
	claimCalls     []claimCall
	finishCalls    []finishCall
	subscribeCalls []subscribeCall
}

func newRepository(subscriptionCount int, nodes ...*adagio.Node) repository {
	wg := sync.WaitGroup{}
	wg.Add(subscriptionCount)

	repo := repository{
		nodes:             map[string]*adagio.Node{},
		subscriptionCount: wg,
	}

	for _, node := range nodes {
		repo.nodes[node.Spec.Name] = node
	}

	return repo
}

type claimCall struct {
	runID, name string
}

func claims(count int, runID, name string) (calls []claimCall) {
	calls = make([]claimCall, 0, count)
	for i := 0; i < count; i++ {
		calls = append(calls, claimCall{runID, name})
	}

	return
}

func (r *repository) ClaimNode(runID string, name string) (*adagio.Node, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// track claim called
	r.claimCalls = append(r.claimCalls, claimCall{runID, name})

	// get node if claimable
	node, ok := r.nodes[name]

	// one claim per node
	delete(r.nodes, name)

	return node, ok, nil
}

type finishCall struct {
	runID, name string
	result      *adagio.Node_Result
}

func (r *repository) FinishNode(runID string, name string, result *adagio.Node_Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.finishCalls = append(r.finishCalls, finishCall{runID, name, result})

	return nil
}

type subscribeCall struct {
	events chan<- *adagio.Event
	types  []adagio.Event_Type
}

func (r *repository) Subscribe(events chan<- *adagio.Event, types ...adagio.Event_Type) error {
	r.mu.Lock()
	defer func() {
		r.subscriptionCount.Done()

		r.mu.Unlock()
	}()

	r.subscribeCalls = append(r.subscribeCalls, subscribeCall{events, types})

	return nil
}
