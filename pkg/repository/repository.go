package repository

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/georgemac/adagio/internal/controlplaneservice"
	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/worker"
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
	ExampleGraph.Connect(nodeC, nodeD)
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

		t.Run("5 claims are successfully attempted for node a", func(t *testing.T) {
			claimed := attemptNClaims(t, repo, run, nodeA, 5)

			t.Run("only 1 succeeds", func(t *testing.T) {
				assert.Equal(t, int32(1), claimed)
			})
		})
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
