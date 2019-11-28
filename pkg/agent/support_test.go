package agent

import (
	"context"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

type runtime struct {
	name        string
	newFunction func() Function
}

func (r runtime) Name() string { return r.name }

func (r runtime) NewFunction() Function { return r.newFunction() }

type function struct {
	run func(context.Context, *adagio.Node) (*adagio.Result, error)
}

func (c function) Run(ct context.Context, n *adagio.Node) (*adagio.Result, error) {
	return c.run(ct, n)
}

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
	claim       *adagio.Claim
}

func claims(count int, runID, name string, claim *adagio.Claim) (calls []claimCall) {
	calls = make([]claimCall, 0, count)
	for i := 0; i < count; i++ {
		calls = append(calls, claimCall{runID, name, claim})
	}

	return
}

func (r *repository) ClaimNode(_ context.Context, runID string, name string, claim *adagio.Claim) (*adagio.Node, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// track claim called
	r.claimCalls = append(r.claimCalls, claimCall{runID, name, claim})

	// get node if claimable
	node, ok := r.nodes[name]

	// one claim per node
	delete(r.nodes, name)

	return node, ok, nil
}

type finishCall struct {
	runID, name string
	result      *adagio.Node_Result
	claim       *adagio.Claim
}

func (r *repository) FinishNode(_ context.Context, runID string, name string, result *adagio.Node_Result, claim *adagio.Claim) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.finishCalls = append(r.finishCalls, finishCall{runID, name, result, claim})

	return nil
}

type subscribeCall struct {
	agent  *adagio.Agent
	events chan<- *adagio.Event
	types  []adagio.Event_Type
}

func (r *repository) UnsubscribeAll(context.Context, *adagio.Agent, chan<- *adagio.Event) error {
	panic("not implemented")
}

func (r *repository) Subscribe(_ context.Context, agent *adagio.Agent, events chan<- *adagio.Event, types ...adagio.Event_Type) error {
	r.mu.Lock()
	defer func() {
		r.subscriptionCount.Done()

		r.mu.Unlock()
	}()

	r.subscribeCalls = append(r.subscribeCalls, subscribeCall{agent, events, types})

	return nil
}
