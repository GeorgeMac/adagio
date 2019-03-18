package repository

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	nodeA = &adagio.Node{Name: "a"}
	nodeB = &adagio.Node{Name: "b"}
	nodeC = &adagio.Node{Name: "c"}
	nodeD = &adagio.Node{Name: "d"}
	nodeE = &adagio.Node{Name: "e"}
	nodeF = &adagio.Node{Name: "f"}
	nodeG = &adagio.Node{Name: "g"}

	ExampleGraph = graph.New(nodeA,
		nodeB,
		nodeC,
		nodeD,
		nodeE,
		nodeF,
		nodeG)
)

func init() {
	ExampleGraph.Connect(nodeA, nodeC)
	ExampleGraph.Connect(nodeA, nodeD)
	ExampleGraph.Connect(nodeB, nodeD)
	ExampleGraph.Connect(nodeB, nodeF)
	ExampleGraph.Connect(nodeC, nodeE)
	ExampleGraph.Connect(nodeD, nodeE)
	ExampleGraph.Connect(nodeE, nodeG)
	ExampleGraph.Connect(nodeF, nodeG)
}

type Repository interface {
	controlplane.Repository
	worker.Repository
}

type UnsubscribeRepository interface {
	Repository
	UnsubscribeAll(chan<- adagio.Event) error
}

func TestHarness(t *testing.T, repo Repository) {
	t.Helper()

	t.Run("a run is created", func(t *testing.T) {
		run, err := repo.StartRun(adagio.NewGraph(ExampleGraph))
		require.Nil(t, err)
		require.NotNil(t, run)

		t.Run("which can be listed", func(t *testing.T) {
			runs, err := repo.ListRuns()
			require.Nil(t, err)

			assert.Len(t, runs, 1)
			assert.Equal(t, run.ID, runs[0].ID)
		})

		// (›) ---> (c)----
		//   \             \
		//    ------v       v
		//         (d) --> (e) --> (g)
		//    ------^               ^
		//   /                     /
		// (›) --> (f) ------------
		testLayer(t, "input layer", repo, run, []*adagio.Node{nodeC, nodeD, nodeE, nodeF, nodeG}, []*adagio.Node{nodeA, nodeB}, []adagio.Event{
			{Node: nodeA, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeA, From: adagio.RunningState, To: adagio.CompletedState},
			{Node: nodeB, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeB, From: adagio.RunningState, To: adagio.CompletedState},
			{Node: nodeC, From: adagio.WaitingState, To: adagio.ReadyState},
			{Node: nodeD, From: adagio.WaitingState, To: adagio.ReadyState},
			{Node: nodeF, From: adagio.WaitingState, To: adagio.ReadyState},
		}...)

		// (✓) ---> (›)----
		//   \             \
		//    ------v       v
		//         (›) --> (e) --> (g)
		//    ------^               ^
		//   /                     /
		// (✓) --> (›) ------------
		testLayer(t, "layer two", repo, run, []*adagio.Node{nodeE, nodeG}, []*adagio.Node{nodeC, nodeD, nodeF}, []adagio.Event{
			{Node: nodeC, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeC, From: adagio.RunningState, To: adagio.CompletedState},
			{Node: nodeD, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeD, From: adagio.RunningState, To: adagio.CompletedState},
			{Node: nodeE, From: adagio.WaitingState, To: adagio.ReadyState},
			{Node: nodeF, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeF, From: adagio.RunningState, To: adagio.CompletedState},
		}...)

		// (✓) ---> (✓)----
		//   \             \
		//    ------v       v
		//         (✓) --> (›) --> (g)
		//    ------^               ^
		//   /                     /
		// (✓) --> (✓) ------------
		testLayer(t, "layer three", repo, run, []*adagio.Node{nodeG}, []*adagio.Node{nodeE}, []adagio.Event{
			{Node: nodeE, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeE, From: adagio.RunningState, To: adagio.CompletedState},
			{Node: nodeG, From: adagio.WaitingState, To: adagio.ReadyState},
		}...)

		// (✓) ---> (✓)----
		//   \             \
		//    ------v       v
		//         (✓) --> (✓) --> (›)
		//    ------^               ^
		//   /                     /
		// (✓) --> (✓) ------------
		testLayer(t, "final layer", repo, run, nil, []*adagio.Node{nodeG}, []adagio.Event{
			{Node: nodeG, From: adagio.ReadyState, To: adagio.RunningState},
			{Node: nodeG, From: adagio.RunningState, To: adagio.CompletedState},
		}...)
	})
}

func testLayer(t *testing.T, name string, repo Repository, run *adagio.Run, notClaimed, claimed []*adagio.Node, expectedEvents ...adagio.Event) {
	t.Helper()

	var (
		events    = make(chan adagio.Event, len(expectedEvents))
		collected = make([]adagio.Event, 0)
		err       = repo.Subscribe(events, adagio.ReadyState, adagio.RunningState, adagio.CompletedState)
	)
	require.Nil(t, err)

	defer func() {
		if urepo, ok := repo.(UnsubscribeRepository); ok {
			urepo.UnsubscribeAll(events)
		}

		close(events)
	}()

	t.Run(name, func(t *testing.T) {
		canNotClaim(t, repo, run, notClaimed...)

		canClaim(t, repo, run, claimed...)
	})

	canFinish(t, repo, run, claimed...)

	for i := 0; i < len(expectedEvents); i++ {
		select {
		case event := <-events:
			event.Run = nil
			collected = append(collected, event)
		default:
		}
	}

	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].Node.Name < collected[j].Node.Name
	})

	if expectedEvents != nil {
		assert.Equal(t, expectedEvents, collected)
	}
}

func canNotClaim(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Helper()

	t.Run("can not claim", func(t *testing.T) {
		t.Parallel()

		for _, node := range nodes {
			t.Run(fmt.Sprintf("node %q cannot be claimed because it is not ready", node), func(t *testing.T) {
				func(node *adagio.Node) {
					t.Parallel()

					_, err := repo.ClaimNode(run, node)
					assert.Equal(t, adagio.ErrNodeNotReady, errors.Cause(err))
				}(node)
			})
		}
	})
}

func canClaim(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Helper()

	t.Run("can claim", func(t *testing.T) {
		t.Parallel()

		for _, node := range nodes {
			t.Run(fmt.Sprintf("5 concurrent claim attempts for node %q", node), func(t *testing.T) {
				func(node *adagio.Node) {
					t.Parallel()

					claimed := attemptNClaims(t, repo, run, node, 5)

					t.Run("only 1 succeeds", func(t *testing.T) {
						assert.Equal(t, int32(1), claimed)
					})
				}(node)
			})
		}
	})
}

func canFinish(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Helper()

	t.Run("can finish", func(t *testing.T) {
		for _, node := range nodes {
			t.Run(fmt.Sprintf("node %q", node), func(t *testing.T) {
				func(node *adagio.Node) {
					t.Parallel()

					assert.Nil(t, repo.FinishNode(run, node))
				}(node)
			})
		}
	})
}

func attemptNClaims(t *testing.T, repo Repository, run *adagio.Run, node *adagio.Node, n int) (claimed int32) {
	t.Helper()

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := repo.ClaimNode(run, node)
			require.Nil(t, err)

			if ok {
				atomic.AddInt32(&claimed, 1)
			}
		}()
	}

	wg.Wait()

	return
}
