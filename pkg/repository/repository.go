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

	h = &adagio.Node_Spec{
		Name: "h",
		Retry: map[string]*adagio.Node_Spec_Retry{
			"error": {MaxAttempts: 2},
		},
	}

	i = &adagio.Node_Spec{
		Name: "i",
		Retry: map[string]*adagio.Node_Spec_Retry{
			"fail": {MaxAttempts: 2},
		},
	}

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

type Orphaner interface {
	Orphan(run *adagio.Run, spec *adagio.Node_Spec)
}

type OrphanFunc func(*adagio.Run, *adagio.Node_Spec)

func (o OrphanFunc) Orphan(r *adagio.Run, s *adagio.Node_Spec) { o(r, s) }

type UnsubscribeRepository interface {
	Repository
	UnsubscribeAll(chan<- *adagio.Event) error
}

type Constructor func(func() time.Time) (Repository, Orphaner)

func TestHarness(t *testing.T, repoFn Constructor) {
	t.Helper()

	repo, orphaner := repoFn(func() time.Time {
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
					{RunID: run.Id, NodeSpec: c, Type: adagio.Event_NODE_READY},
					{RunID: run.Id, NodeSpec: d, Type: adagio.Event_NODE_READY},
					{RunID: run.Id, NodeSpec: f, Type: adagio.Event_NODE_READY},
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
					{RunID: run.Id, NodeSpec: e, Type: adagio.Event_NODE_READY},
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
					{RunID: run.Id, NodeSpec: g, Type: adagio.Event_NODE_READY},
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
				Events: []*adagio.Event{},
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
				completed(a, nil, success("a")),
				completed(b, nil, success("b")),
				completed(c, map[string][]byte{
					"a": []byte("a"),
				}, success("c")),
				completed(d, map[string][]byte{
					"a": []byte("a"),
					"b": []byte("b"),
				}, success("d")),
				completed(e, map[string][]byte{
					"c": []byte("c"),
					"d": []byte("d"),
				}, success("e")),
				completed(f, map[string][]byte{
					"b": []byte("b"),
				}, success("f")),
				completed(g, map[string][]byte{
					"e": []byte("e"),
					"f": []byte("f"),
				}, success("g")),
			}, runs[0].Nodes)
		})
	})

	t.Run("a run with a failure", func(t *testing.T) {
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
					{RunID: run.Id, NodeSpec: c, Type: adagio.Event_NODE_READY},
					{RunID: run.Id, NodeSpec: d, Type: adagio.Event_NODE_READY},
					{RunID: run.Id, NodeSpec: f, Type: adagio.Event_NODE_READY},
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
				completed(a, nil, success("a")),
				completed(b, nil, success("b")),
				completed(c, map[string][]byte{
					"a": []byte("a"),
				}, success("c")),
				completed(d, map[string][]byte{
					"a": []byte("a"),
					"b": []byte("b"),
				}, fail("d")),
				completed(e, map[string][]byte{
					"c": []byte("c"),
				}),
				completed(f, map[string][]byte{
					"b": []byte("b"),
				}, success("f")),
				completed(g, map[string][]byte{
					"f": []byte("f"),
				}),
			}, runs[1].Nodes)
		})
	})

	t.Run("a run with retries", func(t *testing.T) {
		// (a) --> (h) --> (c)
		//                  ^
		//                 /
		//         (b) ----
		run, err := repo.StartRun(&adagio.GraphSpec{
			Nodes: []*adagio.Node_Spec{
				a,
				b,
				c,
				h,
			},
			Edges: []*adagio.Edge{
				{Source: a.Name, Destination: h.Name},
				{Source: b.Name, Destination: c.Name},
				{Source: h.Name, Destination: c.Name},
			},
		})
		require.Nil(t, err)
		require.NotNil(t, run)

		assert.Equal(t, adagio.Run_WAITING, run.Status)

		for _, layer := range []TestLayer{
			{
				// (›) --> (h) --> (c)
				//                  ^
				//                 /
				//         (›) ----
				Name:        "input layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c", "h"},
				Claimable: map[string]*adagio.Node{
					"a": running(a, nil),
					"b": running(b, nil),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"a": adagio.Node_Result_SUCCESS,
					"b": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: h, Type: adagio.Event_NODE_READY},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (›) --> (›) --> (c)
				//                  ^
				//                 /
				//         (›) ----
				Name:        "retriable error layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c"},
				Claimable: map[string]*adagio.Node{
					"h": running(h, map[string][]byte{
						"a": []byte("a"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"h": adagio.Node_Result_ERROR,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: h, Type: adagio.Event_NODE_READY},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) --> (!) --> (c)
				//                  ^
				//                 /
				//         (✓) ----
				Name:        "successful second attempt layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c"},
				Claimable: map[string]*adagio.Node{
					"h": running(h, map[string][]byte{
						"a": []byte("a"),
					}, errorResult("h")),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"h": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: c, Type: adagio.Event_NODE_READY},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (✓) --> (✓) --> (›)
				//                  ^
				//                 /
				//         (✓) ----
				Name:        "final layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{},
				Claimable: map[string]*adagio.Node{
					"c": running(c, map[string][]byte{
						"b": []byte("b"),
						"h": []byte("h"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"c": adagio.Node_Result_SUCCESS,
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
			assert.Len(t, runs, 3)
			assert.Equal(t, run.Id, runs[2].Id)

			assert.Equal(t, []*adagio.Node{
				completed(a, nil, success("a")),
				completed(b, nil, success("b")),
				completed(c, map[string][]byte{
					"b": []byte("b"),
					"h": []byte("h"),
				}, success("c")),
				completed(h, map[string][]byte{
					"a": []byte("a"),
				}, errorResult("h"), success("h")),
			}, runs[2].Nodes)
		})
	})

	t.Run("a run with exhausted retries", func(t *testing.T) {
		// (a) --> (i) --> (c)
		//                  ^
		//                 /
		//         (b) ----
		run, err := repo.StartRun(&adagio.GraphSpec{
			Nodes: []*adagio.Node_Spec{
				a,
				b,
				c,
				i,
			},
			Edges: []*adagio.Edge{
				{Source: a.Name, Destination: i.Name},
				{Source: b.Name, Destination: c.Name},
				{Source: i.Name, Destination: c.Name},
			},
		})
		require.Nil(t, err)
		require.NotNil(t, run)

		assert.Equal(t, adagio.Run_WAITING, run.Status)

		for _, layer := range []TestLayer{
			{
				// (›) --> (i) --> (c)
				//                  ^
				//                 /
				//         (›) ----
				Name:        "input layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c", "i"},
				Claimable: map[string]*adagio.Node{
					"a": running(a, nil),
					"b": running(b, nil),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"a": adagio.Node_Result_SUCCESS,
					"b": adagio.Node_Result_SUCCESS,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: i, Type: adagio.Event_NODE_READY},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (›) --> (›) --> (c)
				//                  ^
				//                 /
				//         (›) ----
				Name:        "retriable fail layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c"},
				Claimable: map[string]*adagio.Node{
					"i": running(i, map[string][]byte{
						"a": []byte("a"),
					}),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"i": adagio.Node_Result_FAIL,
				},
				Events: []*adagio.Event{
					{RunID: run.Id, NodeSpec: i, Type: adagio.Event_NODE_READY},
				},
				RunStatus: adagio.Run_RUNNING,
			},
			{
				// (›) --> (✗) --> (.)
				//                  ^
				//                 /
				//         (›) ----
				Name:        "second retriable fail layer",
				Repository:  repo,
				Run:         run,
				Unclaimable: []string{"c"},
				Claimable: map[string]*adagio.Node{
					"i": running(i, map[string][]byte{
						"a": []byte("a"),
					}, fail("i")),
				},
				Finish: map[string]adagio.Node_Result_Conclusion{
					"i": adagio.Node_Result_FAIL,
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
			assert.Len(t, runs, 4)
			assert.Equal(t, run.Id, runs[3].Id)

			assert.Equal(t, []*adagio.Node{
				completed(a, nil, success("a")),
				completed(b, nil, success("b")),
				completed(c, map[string][]byte{
					"b": []byte("b"),
				}),
				completed(i, map[string][]byte{
					"a": []byte("a"),
				}, fail("i"), fail("i")),
			}, runs[3].Nodes)
		})
	})

	t.Run("a run with an orphaned node", func(t *testing.T) {
		// (a)
		run, err := repo.StartRun(&adagio.GraphSpec{
			Nodes: []*adagio.Node_Spec{
				a,
			},
			Edges: []*adagio.Edge{},
		})
		require.Nil(t, err)
		require.NotNil(t, run)

		var (
			events    = make(chan *adagio.Event, 5)
			collected = make([]*adagio.Event, 0)
		)

		err = repo.Subscribe(events, adagio.Event_NODE_READY, adagio.Event_NODE_ORPHANED)
		require.Nil(t, err)

		assert.Equal(t, adagio.Run_WAITING, run.Status)

		t.Run("an initial claim is made", func(t *testing.T) {
			// ensure node can initially be claimed
			canClaim(t, repo, run, map[string]*adagio.Node{"a": running(a, nil)})
		})

		// orphan node "a" from run
		orphaner.Orphan(run, a)

		t.Run("the orphaned node", func(t *testing.T) {
			// ensure orphaned node can be claimed again and has no results yet
			canClaim(t, repo, run, map[string]*adagio.Node{"a": running(a, nil)})
		})

		// can error the node
		canFinish(t, repo, run, map[string]adagio.Node_Result_Conclusion{"a": adagio.Node_Result_ERROR})

		select {
		case event := <-events:
			collected = append(collected, event)
		case <-time.After(5 * time.Second):
			t.Error("timeout collecting events")
			return
		}

		expected := []*adagio.Event{
			{RunID: run.Id, NodeSpec: a, Type: adagio.Event_NODE_ORPHANED},
		}
		if !assert.Equal(t, expected, collected) {
			fmt.Println(pretty.Diff(expected, collected))
		}

		t.Run("the run is listed", func(t *testing.T) {
			runs, err := repo.ListRuns()
			require.Nil(t, err)

			// the run is listed
			assert.Len(t, runs, 5)
			assert.Equal(t, run.Id, runs[4].Id)

			assert.Equal(t, []*adagio.Node{
				completed(a, nil, errorResult("a")),
			}, runs[4].Nodes)
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
		err       = l.Repository.Subscribe(events, adagio.Event_NODE_READY)
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

func running(spec *adagio.Node_Spec, inputs map[string][]byte, attempts ...*adagio.Node_Result) *adagio.Node {
	n := node(spec, adagio.Node_RUNNING, inputs)
	n.StartedAt = when.Format(time.RFC3339)
	n.Attempts = attempts
	return n
}

func completed(spec *adagio.Node_Spec, inputs map[string][]byte, attempts ...*adagio.Node_Result) *adagio.Node {
	n := node(spec, adagio.Node_COMPLETED, inputs)
	n.Attempts = attempts

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
	return result(output, adagio.Node_Result_ERROR)
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
