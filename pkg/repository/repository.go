package repository

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	a = &adagio.Node_Spec{Name: "a"}
	b = &adagio.Node_Spec{Name: "b"}
	c = &adagio.Node_Spec{Name: "c"}
	d = &adagio.Node_Spec{Name: "d"}
	e = &adagio.Node_Spec{Name: "e"}
	f = &adagio.Node_Spec{Name: "f"}
	g = &adagio.Node_Spec{Name: "g"}

	when = time.Date(2019, 5, 24, 8, 2, 0, 0, time.UTC)

	ExampleGraph = &adagio.GraphSpec{
		Nodes: []*adagio.Node_Spec{
			a,
			b,
			c,
			d,
			e,
			f,
			g,
		},
		Edges: []*adagio.Edge{
			{Source: a.Name, Destination: c.Name},
			{Source: a.Name, Destination: d.Name},
			{Source: b.Name, Destination: d.Name},
			{Source: b.Name, Destination: f.Name},
			{Source: c.Name, Destination: e.Name},
			{Source: d.Name, Destination: e.Name},
			{Source: e.Name, Destination: g.Name},
			{Source: f.Name, Destination: g.Name},
		},
	}
)

type Repository interface {
	controlplane.Repository
	worker.Repository
}

type UnsubscribeRepository interface {
	Repository
	UnsubscribeAll(chan<- *adagio.Event) error
}

type Constructor func(func() time.Time) Repository

func TestHarness(t *testing.T, repoFn Constructor) {
	t.Helper()

	repo := repoFn(func() time.Time {
		return when
	})

	t.Run("a run is created", func(t *testing.T) {
		run, err := repo.StartRun(ExampleGraph)
		require.Nil(t, err)
		require.NotNil(t, run)

		assert.Equal(t, adagio.Run_WAITING, run.Status)

		for _, layer := range []TestLayer{
			{
				// (›) ---> (c)----
				//   \             \
				//    ------v       v
				//         (d) --> (e) --> (g)
				//    ------^               ^
				//   /                     /
				// (›) --> (f) ------------
				Name:        "input layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c", "d", "e", "f", "g"},
				Claimable: map[string]*adagio.Node{
					"a": running(a, nil),
					"b": running(b, nil),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"a": adagio.Node_Result_SUCCESS,
					"b": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "a"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "a"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "b"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "b"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) ---> (›)----
				//   \             \
				//    ------v       v
				//         (›) --> (e) --> (g)
				//    ------^               ^
				//   /                     /
				// (✓) --> (›) ------------
				Name:        "second layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"e", "g"},
				Claimable: map[string]*adagio.Node{
					"c": running(c, map[string][]byte{
						"a": []byte("a"),
					}),
					"d": running(d, map[string][]byte{
						"a": []byte("a"),
						"b": []byte("b"),
					}),
					"f": running(f, map[string][]byte{
						"b": []byte("b"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"c": adagio.Node_Result_SUCCESS,
					"d": adagio.Node_Result_SUCCESS,
					"f": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "e"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) ---> (✓)----
				//   \             \
				//    ------v       v
				//         (✓) --> (›) --> (g)
				//    ------^               ^
				//   /                     /
				// (✓) --> (✓) ------------
				Name:        "third layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"g"},
				Claimable: map[string]*adagio.Node{
					"e": running(e, map[string][]byte{
						"c": []byte("c"),
						"d": []byte("d"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"e": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "e"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "e"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "g"}, Type: adagio.Event_STATE_TRANSITION},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) ---> (✓)----
				//   \             \
				//    ------v       v
				//         (✓) --> (✓) --> (›)
				//    ------^               ^
				//   /                     /
				// (✓) --> (✓) ------------
				Name:       "final layer",
				Repository: repo,
				Run:        run,
				Claimable: map[string]*adagio.Node{
					"g": running(g, map[string][]byte{
						"e": []byte("e"),
						"f": []byte("f"),
					}),
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "g"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "g"}, Type: adagio.Event_STATE_TRANSITION},
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"g": adagio.Node_Result_SUCCESS,
				},
				RunStatus: adagio.Run_COMPLETED,
			},
		} {
			layer.Exec(t)
		}

		t.Run("the run is listed", func(t *testing.T) {
			runs, err := repo.ListRuns()
			require.Nil(t, err)

			assert.Len(t, runs, 1)
			assert.Equal(t, run.Id, runs[0].Id)

			assert.Equal(t, []*adagio.Node{
				completed(a, success("a"), nil),
				completed(b, success("b"), nil),
				completed(c, success("c"), map[string][]byte{
					"a": []byte("a"),
				}),
				completed(d, success("d"), map[string][]byte{
					"a": []byte("a"),
					"b": []byte("b"),
				}),
				completed(e, success("e"), map[string][]byte{
					"c": []byte("c"),
					"d": []byte("d"),
				}),
				completed(f, success("f"), map[string][]byte{
					"b": []byte("b"),
				}),
				completed(g, success("g"), map[string][]byte{
					"e": []byte("e"),
					"f": []byte("f"),
				}),
			}, runs[0].Nodes)
		})
	})

	t.Run("a second run is created", func(t *testing.T) {
		run, err := repo.StartRun(ExampleGraph)
		require.Nil(t, err)
		require.NotNil(t, run)

		assert.Equal(t, adagio.Run_WAITING, run.Status)

		for _, layer := range []TestLayer{
			{
				// (›) ---> (c)----
				//   \             \
				//    ------v       v
				//         (d) --> (e) --> (g)
				//    ------^               ^
				//   /                     /
				// (›) --> (f) ------------
				Name:        "input layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c", "d", "e", "f", "g"},
				Claimable: map[string]*adagio.Node{
					"a": running(a, nil),
					"b": running(b, nil),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"a": adagio.Node_Result_SUCCESS,
					"b": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "a"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "a"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "b"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "b"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) ---> (›)----
				//   \             \
				//    ------v       v
				//         (✗) --> (.) --> (.)
				//    ------^               ^
				//   /                     /
				// (✓) --> (›) ------------
				Name:        "second layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"e", "g"},
				Claimable: map[string]*adagio.Node{
					"c": running(c, map[string][]byte{
						"a": []byte("a"),
					}),
					"d": running(d, map[string][]byte{
						"a": []byte("a"),
						"b": []byte("b"),
					}),
					"f": running(f, map[string][]byte{
						"b": []byte("b"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"c": adagio.Node_Result_SUCCESS,
					"d": adagio.Node_Result_FAIL,
					"f": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "c"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "d"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "e"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "f"}, Type: adagio.Event_STATE_TRANSITION},
					{RunID: run.Id, NodeSpec: &adagio.Node_Spec{Name: "g"}, Type: adagio.Event_STATE_TRANSITION},
				},
				RunStatus: adagio.Run_COMPLETED,
			},
		} {
			layer.Exec(t)
		}

		t.Run("the run is listed", func(t *testing.T) {
			runs, err := repo.ListRuns()
			require.Nil(t, err)

			// the run is listed
			assert.Len(t, runs, 2)
			assert.Equal(t, run.Id, runs[1].Id)

			assert.Equal(t, []*adagio.Node{
				completed(a, success("a"), nil),
				completed(b, success("b"), nil),
				completed(c, success("c"), map[string][]byte{
					"a": []byte("a"),
				}),
				completed(d, fail("d"), map[string][]byte{
					"a": []byte("a"),
					"b": []byte("b"),
				}),
				completed(e, nil, map[string][]byte{
					"c": []byte("c"),
				}),
				completed(f, success("f"), map[string][]byte{
					"b": []byte("b"),
				}),
				completed(g, nil, map[string][]byte{
					"f": []byte("f"),
				}),
			}, runs[1].Nodes)
		})
	})
}

type TestLayer struct {
	Name        string
	Repository  Repository
	Run         *adagio.Run
	Unclaimable []string
	Claimable   map[string]*adagio.Node
	Finish      map[string]adagio.Node_Result_Conclusion
	Events      []*adagio.Event
	RunStatus   adagio.Run_Status
}

func (l *TestLayer) Exec(t *testing.T) {
	t.Helper()

	var (
		events    = make(chan *adagio.Event, len(l.Events))
		collected = make([]*adagio.Event, 0)
		err       = l.Repository.Subscribe(events, adagio.Node_READY, adagio.Node_RUNNING, adagio.Node_COMPLETED)
	)
	require.Nil(t, err)

	defer func() {
		if urepo, ok := l.Repository.(UnsubscribeRepository); ok {
			urepo.UnsubscribeAll(events)
		}

		close(events)
	}()

	t.Run(l.Name, func(t *testing.T) {
		canNotClaim(t, l.Repository, l.Run, l.Unclaimable...)

		canClaim(t, l.Repository, l.Run, l.Claimable)
	})

	canFinish(t, l.Repository, l.Run, l.Finish)

	t.Run(fmt.Sprintf("the run is reported with a status of %q", l.RunStatus), func(t *testing.T) {
		// check run reports expected status
		run, err := l.Repository.InspectRun(l.Run.Id)
		require.Nil(t, err)
		require.Equal(t, l.RunStatus, run.Status)
	})

	for i := 0; i < len(l.Events); i++ {
		select {
		case event := <-events:
			collected = append(collected, event)
		case <-time.After(5 * time.Second):
			t.Error("timeout collecting events")
			return
		}
	}

	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].NodeSpec.Name < collected[j].NodeSpec.Name
	})

	if l.Events != nil {
		if !assert.Equal(t, l.Events, collected) {
			fmt.Println(pretty.Diff(l.Events, collected))
		}
	}
}

func canNotClaim(t *testing.T, repo Repository, run *adagio.Run, names ...string) {
	t.Helper()

	t.Run("can not claim", func(t *testing.T) {
		t.Parallel()

		for _, name := range names {
			t.Run(fmt.Sprintf("node %q cannot be claimed because it is not ready", name), func(t *testing.T) {
				func(name string) {
					t.Parallel()

					_, _, err := repo.ClaimNode(run.Id, name)
					assert.Equal(t, adagio.ErrNodeNotReady, errors.Cause(err))
				}(name)
			})
		}
	})
}

func canClaim(t *testing.T, repo Repository, run *adagio.Run, nodes map[string]*adagio.Node) {
	t.Helper()

	t.Run("can claim", func(t *testing.T) {
		t.Parallel()

		for name, node := range nodes {
			t.Run(fmt.Sprintf("5 concurrent claim attempts for node %q", name), func(t *testing.T) {
				func(name string, node *adagio.Node) {
					t.Parallel()

					claimed := attemptNClaims(t, repo, run, name, 5)

					t.Run("and it returns the correct node", func(t *testing.T) {
						assert.Equal(t, node, claimed)
					})
				}(name, node)
			})
		}
	})
}

func canFinish(t *testing.T, repo Repository, run *adagio.Run, names map[string]adagio.Node_Result_Conclusion) {
	t.Helper()

	t.Run("can finish", func(t *testing.T) {
		for name, conclusion := range names {
			t.Run(fmt.Sprintf("node %q", name), func(t *testing.T) {
				func(name string, conclusion adagio.Node_Result_Conclusion) {
					t.Parallel()

					assert.Nil(t, repo.FinishNode(run.Id, name, &adagio.Node_Result{
						Conclusion: conclusion,
						Output:     []byte(name),
					}))
				}(name, conclusion)
			})
		}
	})
}

func attemptNClaims(t *testing.T, repo Repository, run *adagio.Run, name string, n int) (node *adagio.Node) {
	t.Helper()

	t.Run("only one successful claim is made", func(t *testing.T) {
		var (
			wg    sync.WaitGroup
			count int32
		)

		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				cnode, ok, err := repo.ClaimNode(run.Id, name)
				require.Nil(t, err)

				if ok {
					atomic.AddInt32(&count, 1)
					node = cnode
					return
				}

				require.Nil(t, cnode)
			}()
		}

		wg.Wait()

		assert.Equal(t, int32(1), count)
	})

	return
}

func waiting(spec *adagio.Node_Spec) *adagio.Node {
	return node(spec, adagio.Node_WAITING, nil)
}

func ready(spec *adagio.Node_Spec, inputs map[string][]byte) *adagio.Node {
	return node(spec, adagio.Node_READY, inputs)
}

func running(spec *adagio.Node_Spec, inputs map[string][]byte) *adagio.Node {
	n := node(spec, adagio.Node_RUNNING, inputs)
	n.StartedAt = when.Format(time.RFC3339)
	return n
}

func completed(spec *adagio.Node_Spec, result *adagio.Node_Result, inputs map[string][]byte) *adagio.Node {
	n := node(spec, adagio.Node_COMPLETED, inputs)
	if result != nil {
		n.Attempts = []*adagio.Node_Result{result}
	}

	n.StartedAt = when.Format(time.RFC3339)
	n.FinishedAt = when.Format(time.RFC3339)
	return n
}

func node(spec *adagio.Node_Spec, status adagio.Node_Status, inputs map[string][]byte) *adagio.Node {
	return &adagio.Node{Spec: spec, Status: status, Inputs: inputs}
}

func success(output string) *adagio.Node_Result {
	return result(output, adagio.Node_Result_SUCCESS)
}

func fail(output string) *adagio.Node_Result {
	return result(output, adagio.Node_Result_FAIL)
}

func none() *adagio.Node_Result { return result("", adagio.Node_Result_NONE) }

func errorResult(output string) *adagio.Node_Result {
	return result(output, adagio.Node_Result_FAIL)
}

func result(output string, conclusion adagio.Node_Result_Conclusion) *adagio.Node_Result {
	var data []byte
	if output != "" {
		data = []byte(output)
	}

	return &adagio.Node_Result{
		Conclusion: conclusion,
		Output:     data,
	}
}
