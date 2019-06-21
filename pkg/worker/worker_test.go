package worker

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_HappyPath(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "test",
			},
		}

		runtimeCalls uint64

		runtimes = map[string]Runtime{
			"test": RuntimeFunc(func(n *adagio.Node) (*adagio.Result, error) {
				require.Equal(t, n, node)

				atomic.AddUint64(&runtimeCalls, 1)

				return &adagio.Result{
					Conclusion: adagio.Conclusion_SUCCESS,
				}, nil
			}),
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		pool = NewPool(&repo, runtimes, WorkerCount(5))

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

		assert.Equal(t, []adagio.Node_Status{adagio.Node_READY}, call.statuses)

		// feed each subscriber an event for node "foo"
		call.events <- &adagio.Event{RunID: "bar", NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "test"}}
	}

	// stop running
	cancel()
	<-done

	// ensure 5 claims are attempted for run "bar" node "foo"
	assert.Equal(t, claims(5, "bar", "foo"), repo.claimCalls)

	// ensure 1 finish call is made
	require.Len(t, repo.finishCalls, 1)
	assert.Equal(t, finishCall{"bar", "foo", &adagio.Result{
		Conclusion: adagio.Conclusion_SUCCESS,
	}}, repo.finishCalls[0])

	// ensure runtime was invoked once
	assert.Equal(t, uint64(1), runtimeCalls)
}

func TestPool_Error_RuntimeDoesNotExist(t *testing.T) {
	var (
		node = &adagio.Node{
			Spec: &adagio.Node_Spec{
				Name:    "foo",
				Runtime: "unknown",
			},
		}

		runtimeCalls uint64

		runtimes = map[string]Runtime{
			"known": RuntimeFunc(func(n *adagio.Node) (*adagio.Result, error) {
				atomic.AddUint64(&runtimeCalls, 1)

				return &adagio.Result{
					Conclusion: adagio.Conclusion_ERROR,
				}, nil
			}),
		}

		// new repository which expects 5 subscriptions
		repo = newRepository(5, node)

		pool = NewPool(&repo, runtimes, WorkerCount(5))

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

		assert.Equal(t, []adagio.Node_Status{adagio.Node_READY}, call.statuses)

		// feed each subscriber an event for node "foo"
		call.events <- &adagio.Event{RunID: "bar", NodeSpec: &adagio.Node_Spec{Name: "foo", Runtime: "test"}}
	}

	// stop running
	cancel()
	<-done

	// ensure no claims were attempted
	require.Nil(t, repo.claimCalls)

	// ensure no finish calls were made
	require.Nil(t, repo.finishCalls)

	// ensure runtime was never invoked
	assert.Equal(t, uint64(0), runtimeCalls)
}
