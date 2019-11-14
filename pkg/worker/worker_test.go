package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_HappyPath_NODE_READY(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "test",
			},
		}

		parseCalls uint64
		runCalls   uint64

		runtimes = map[string]func() Runtime{
			"test": func() Runtime {
				return runtime{
					parse: func(n *adagio.Node) error {
						atomic.AddUint64(&parseCalls, 1)

						require.Equal(t, n, node)

						return nil
					},
					run: func() (*adagio.Result, error) {
						atomic.AddUint64(&runCalls, 1)

						return &adagio.Result{
							Conclusion: adagio.Result_SUCCESS,
						}, nil
					},
				}
			},
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		claim     = &adagio.Claim{Id: "claim"}
		claimFunc = func() Claimer {
			return ClaimerFunc(func() *adagio.Claim {
				return claim
			})
		}

		pool = NewPool(&repo, runtimes, WithWorkerCount(5), WithClaimerFunc(claimFunc))

		done         = make(chan struct{})
		ctxt, cancel = context.WithCancel(context.Background())
	)

	go func() {
		pool.Run(ctxt)
		done <- struct{}{}
	}()

	// wait for all subscriptions
	repo.subscriptionCount.Wait()

	// ensure we have five repo calls
	require.Len(t, repo.subscribeCalls, 5)

	// ensure each subscriber supplied a channel for ready events
	for _, call := range repo.subscribeCalls {
		require.NotNil(t, call.agent)
		require.NotNil(t, call.events)

		assert.Equal(t, []adagio.Event_Type{
			adagio.Event_NODE_READY,
			adagio.Event_NODE_ORPHANED,
		}, call.types)

		// feed each subscriber an event for node "foo"
		call.events <- &adagio.Event{
			RunID:    "bar",
			NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "test"},
			Type:     adagio.Event_NODE_READY,
		}
	}

	// stop running
	cancel()
	<-done

	// ensure 5 claims are attempted for run "bar" node "foo"
	assert.Equal(t, claims(5, "bar", "foo", claim), repo.claimCalls)

	// ensure 1 finish call is made
	require.Len(t, repo.finishCalls, 1)
	assert.Equal(t, finishCall{"bar", "foo", &adagio.Node_Result{
		Conclusion: adagio.Node_Result_SUCCESS,
	}, claim}, repo.finishCalls[0])

	// ensure runtime was invoked once
	assert.Equal(t, uint64(1), parseCalls)
	assert.Equal(t, uint64(1), runCalls)
}

func TestPool_Error_RuntimeDoesNotExist(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "unknown",
			},
		}

		runCalls   uint64
		parseCalls uint64

		runtimes = map[string]func() Runtime{
			"known": func() Runtime {
				return runtime{
					parse: func(n *adagio.Node) error {
						atomic.AddUint64(&parseCalls, 1)

						require.Equal(t, n, node)

						return nil
					},
					run: func() (*adagio.Result, error) {
						atomic.AddUint64(&runCalls, 1)

						return &adagio.Result{
							Conclusion: adagio.Result_SUCCESS,
						}, nil
					},
				}
			},
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		claim     = &adagio.Claim{Id: "claim"}
		claimFunc = func() Claimer {
			return ClaimerFunc(func() *adagio.Claim {
				return claim
			})
		}
		pool = NewPool(&repo, runtimes, WithWorkerCount(5), WithClaimerFunc(claimFunc))

		done         = make(chan struct{})
		ctxt, cancel = context.WithCancel(context.Background())
	)

	go func() {
		pool.Run(ctxt)
		done <- struct{}{}
	}()

	// wait for all subscriptions
	repo.subscriptionCount.Wait()

	// ensure we have five repo calls
	require.Len(t, repo.subscribeCalls, 5)

	// ensure each subscriber supplied a channel for ready events
	for _, call := range repo.subscribeCalls {
		require.NotNil(t, call.events)

		assert.Equal(t, []adagio.Event_Type{
			adagio.Event_NODE_READY,
			adagio.Event_NODE_ORPHANED,
		}, call.types)

		// feed each subscriber an event for node "foo"
		call.events <- &adagio.Event{
			RunID:    "bar",
			NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "test"},
			Type:     adagio.Event_NODE_READY,
		}
	}

	// stop running
	cancel()
	<-done

	// ensure no claims were attempted
	require.Nil(t, repo.claimCalls)

	// ensure no finish calls were made
	require.Nil(t, repo.finishCalls)

	// ensure runtime was never invoked
	assert.Equal(t, uint64(0), runCalls)
	assert.Equal(t, uint64(0), parseCalls)
}

func TestPool_Error_RuntimeError(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "error",
			},
		}

		runCalls   uint64
		parseCalls uint64

		runtimes = map[string]func() Runtime{
			"error": func() Runtime {
				return runtime{
					parse: func(n *adagio.Node) error {
						atomic.AddUint64(&parseCalls, 1)

						require.Equal(t, n, node)

						return nil
					},
					run: func() (*adagio.Result, error) {
						atomic.AddUint64(&runCalls, 1)

						return nil, errors.New("something went wrong")
					},
				}
			},
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		claim     = &adagio.Claim{Id: "claim"}
		claimFunc = func() Claimer {
			return ClaimerFunc(func() *adagio.Claim {
				return claim
			})
		}
		pool = NewPool(&repo, runtimes, WithWorkerCount(5), WithClaimerFunc(claimFunc))

		done         = make(chan struct{})
		ctxt, cancel = context.WithCancel(context.Background())
	)

	go func() {
		pool.Run(ctxt)
		done <- struct{}{}
	}()

	// wait for all subscriptions
	repo.subscriptionCount.Wait()

	// ensure we have five repo calls
	require.Len(t, repo.subscribeCalls, 5)

	// ensure each subscriber supplied a channel for ready events
	for _, call := range repo.subscribeCalls {
		require.NotNil(t, call.events)

		assert.Equal(t, []adagio.Event_Type{
			adagio.Event_NODE_READY,
			adagio.Event_NODE_ORPHANED,
		}, call.types)

		// feed each subscriber an event for node "foo"
		call.events <- &adagio.Event{
			RunID:    "bar",
			NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "error"},
			Type:     adagio.Event_NODE_READY,
		}
	}

	// stop running
	cancel()
	<-done

	// ensure 5 claims are attempted for run "bar" node "foo"
	assert.Equal(t, claims(5, "bar", "foo", claim), repo.claimCalls)

	// ensure 1 finish call is made
	require.Len(t, repo.finishCalls, 1)
	assert.Equal(t, finishCall{"bar", "foo", &adagio.Node_Result{
		Output:     []byte("something went wrong"),
		Conclusion: adagio.Node_Result_ERROR,
	}, claim}, repo.finishCalls[0])

	// ensure runtime was never invoked
	assert.Equal(t, uint64(1), runCalls)
	assert.Equal(t, uint64(1), parseCalls)
}

func TestPool_Error_NODE_ORPHANED(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "test",
			},
		}

		runCalls   uint64
		parseCalls uint64

		runtimes = map[string]func() Runtime{
			"test": func() Runtime {
				return runtime{
					parse: func(n *adagio.Node) error {
						atomic.AddUint64(&parseCalls, 1)

						require.Equal(t, n, node)

						return nil
					},
					run: func() (*adagio.Result, error) {
						atomic.AddUint64(&runCalls, 1)

						return &adagio.Result{
							Conclusion: adagio.Result_SUCCESS,
						}, nil
					},
				}
			},
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		claim     = &adagio.Claim{Id: "claim"}
		claimFunc = func() Claimer {
			return ClaimerFunc(func() *adagio.Claim {
				return claim
			})
		}
		pool = NewPool(&repo, runtimes, WithWorkerCount(5), WithClaimerFunc(claimFunc))

		done         = make(chan struct{})
		ctxt, cancel = context.WithCancel(context.Background())
	)

	go func() {
		pool.Run(ctxt)
		done <- struct{}{}
	}()

	// wait for all subscriptions
	repo.subscriptionCount.Wait()

	// ensure we have five repo calls
	require.Len(t, repo.subscribeCalls, 5)

	// ensure each subscriber supplied a channel for ready events
	for _, call := range repo.subscribeCalls {
		require.NotNil(t, call.events)

		assert.Equal(t, []adagio.Event_Type{
			adagio.Event_NODE_READY,
			adagio.Event_NODE_ORPHANED,
		}, call.types)

		// feed each subscriber an orphaned event for node "foo"
		call.events <- &adagio.Event{
			RunID:    "bar",
			NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "test"},
			Type:     adagio.Event_NODE_ORPHANED,
		}
	}

	// stop running
	cancel()
	<-done

	// ensure 5 claims are attempted for run "bar" node "foo"
	assert.Equal(t, claims(5, "bar", "foo", claim), repo.claimCalls)

	// ensure 1 finish call is made
	require.Len(t, repo.finishCalls, 1)
	assert.Equal(t, finishCall{"bar", "foo", &adagio.Node_Result{
		Output:     []byte("node was orphaned"),
		Conclusion: adagio.Node_Result_ERROR,
	}, claim}, repo.finishCalls[0])

	// ensure runtime was never invoked
	assert.Equal(t, uint64(0), runCalls)
	assert.Equal(t, uint64(0), parseCalls)
}
