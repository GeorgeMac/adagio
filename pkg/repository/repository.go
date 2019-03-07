package repository

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/georgemac/adagio/internal/controlplaneservice"
	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
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
	controlplaneservice.Repository
	worker.Repository
}

type UnsubscribeRepository interface {
	Repository
	UnsubscribeAll(chan<- adagio.Event) error
}

func TestHarness(t *testing.T, repo Repository) {
	t.Helper()

	var (
		events    = make(chan adagio.Event, 10)
		collected = make([]adagio.Event, 0)
		err       = repo.Subscribe(events, adagio.ReadyState, adagio.RunningState, adagio.CompletedState)
	)
	require.Nil(t, err)

	go func() {
		defer func() {
			if urepo, ok := repo.(UnsubscribeRepository); ok {
				urepo.UnsubscribeAll(events)
			}

			close(events)
		}()

		for event := range events {
			collected = append(collected, event)
		}
	}()

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
		canNotClaim(t, repo, run, nodeC, nodeD, nodeE, nodeF, nodeG)

		canClaim(t, repo, run, nodeA, nodeB)

		canFinish(t, repo, run, nodeA, nodeB)

		// (✓) ---> (›)----
		//   \             \
		//    ------v       v
		//         (›) --> (e) --> (g)
		//    ------^               ^
		//   /                     /
		// (✓) --> (›) ------------
		canNotClaim(t, repo, run, nodeE, nodeG)

		canClaim(t, repo, run, nodeC, nodeD, nodeF)

		canFinish(t, repo, run, nodeC, nodeD, nodeF)

		// (✓) ---> (✓)----
		//   \             \
		//    ------v       v
		//         (✓) --> (›) --> (g)
		//    ------^               ^
		//   /                     /
		// (✓) --> (✓) ------------
		canNotClaim(t, repo, run, nodeG)

		canClaim(t, repo, run, nodeE)

		canFinish(t, repo, run, nodeE)

		// (✓) ---> (✓)----
		//   \             \
		//    ------v       v
		//         (✓) --> (✓) --> (›)
		//    ------^               ^
		//   /                     /
		// (✓) --> (✓) ------------
		canClaim(t, repo, run, nodeG)

		canFinish(t, repo, run, nodeG)
	})
}

func canNotClaim(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Run("can not claim", func(t *testing.T) {
		for _, node := range nodes {
			t.Run(fmt.Sprintf("node %q cannot be claimed because it is not ready", node), func(t *testing.T) {
				// assigned node locally
				node := node

				t.Parallel()

				_, err := repo.ClaimNode(run, node)
				assert.Equal(t, adagio.ErrNodeNotReady, errors.Cause(err))
			})
		}
	})
}

func canClaim(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Run("can claim", func(t *testing.T) {
		for _, node := range nodes {
			t.Run(fmt.Sprintf("5 concurrent claim attempts for node %q", node), func(t *testing.T) {
				// assigned node locally
				node := node

				t.Parallel()

				claimed := attemptNClaims(t, repo, run, node, 5)

				t.Run("only 1 succeeds", func(t *testing.T) {
					assert.Equal(t, int32(1), claimed)
				})
			})
		}
	})
}

func canFinish(t *testing.T, repo Repository, run *adagio.Run, nodes ...*adagio.Node) {
	t.Helper()

	t.Run("can finish", func(t *testing.T) {
		for _, node := range nodes {
			t.Run(fmt.Sprintf("node %q", node), func(t *testing.T) {
				// assigned node locally
				node := node

				t.Parallel()

				assert.Nil(t, repo.FinishNode(run, node))
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
