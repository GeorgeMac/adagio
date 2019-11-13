package worker

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/oklog/ulid/v2"
)

var (
	// ErrRuntimeDoesNotExist is returned when a node is claimed with an
	// unkown runtime type
	ErrRuntimeDoesNotExist = errors.New("runtime does not exist")
)

// Repository is the minimal interface for a backing repository which can
// notify of node related events, issue node claims and finalize the result
// of executing a node
type Repository interface {
	ClaimNode(runID, name string, claim *adagio.Claim) (*adagio.Node, bool, error)
	FinishNode(runID, name string, result *adagio.Node_Result, claim *adagio.Claim) error
	Subscribe(agent *adagio.Agent, events chan<- *adagio.Event, types ...adagio.Event_Type) error
	UnsubscribeAll(*adagio.Agent, chan<- *adagio.Event) error
}

// Runtime is a type which can execute a node and produce a result
type Runtime interface {
	Run(*adagio.Node) (*adagio.Result, error)
}

// RuntimeFunc is a function which can be used as a Runtime
type RuntimeFunc func(*adagio.Node) (*adagio.Result, error)

// Run delegates to the wrapped RuntimeFunc
func (fn RuntimeFunc) Run(n *adagio.Node) (*adagio.Result, error) { return fn(n) }

// Claimer is used to generate claims
type Claimer interface {
	NewClaim() *adagio.Claim
}

// ClaimerFunc is a function which can be used as a Claimer
type ClaimerFunc func() *adagio.Claim

// NewClaim delegates to underlying ClaimerFunc
func (fn ClaimerFunc) NewClaim() *adagio.Claim { return fn() }

// Pool spins up a number of worker goroutines which subscribe to nodes
// transitioning into the ready state and then attempts to claim and
// process them
type Pool struct {
	repo     Repository
	runtimes map[string]Runtime

	size int

	newClaimer func() Claimer
}

// NewPool constructs and configures a new node pool for execution
func NewPool(repo Repository, runtimes map[string]Runtime, opts ...Option) *Pool {
	pool := &Pool{
		repo:     repo,
		runtimes: runtimes,
		size:     1,
		newClaimer: func() Claimer {
			entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

			return ClaimerFunc(func() *adagio.Claim {
				return &adagio.Claim{
					Id: ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy).String(),
				}
			})
		},
	}

	Options(opts).Apply(pool)

	return pool
}

// Run begins the configured number of workers and responds to cancelation
// of the supplied context
func (p *Pool) Run(ctxt context.Context) {
	var (
		entropy  = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		runtimes = []*adagio.Runtime{}
		wg       sync.WaitGroup
	)

	for runtime := range p.runtimes {
		runtimes = append(runtimes, &adagio.Runtime{Name: runtime})
	}

	sort.Slice(runtimes, func(i, j int) bool {
		return runtimes[i].Name < runtimes[j].Name
	})

	for i := 0; i < p.size; i++ {
		agent := &adagio.Agent{
			Id:       ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy).String(),
			Runtimes: runtimes,
		}

		wg.Add(1)
		go func(agent *adagio.Agent) {
			defer wg.Done()

			var (
				events  = make(chan *adagio.Event, 10)
				claimer = p.newClaimer()
			)

			p.repo.Subscribe(agent, events, adagio.Event_NODE_READY, adagio.Event_NODE_ORPHANED)

			for {
				select {
				case event := <-events:
					if err := p.handleEvent(claimer, event); err != nil {
						log.Println(err)
					}

				case <-ctxt.Done():
					return
				}
			}
		}(agent)
	}

	wg.Wait()
}

func (p *Pool) handleEvent(claimer Claimer, event *adagio.Event) error {
	runtime, ok := p.runtimes[event.NodeSpec.Runtime]
	if !ok {
		return ErrRuntimeDoesNotExist
	}

	// construct a new claim
	claim := claimer.NewClaim()

	node, claimed, err := p.repo.ClaimNode(event.RunID, event.NodeSpec.Name, claim)
	if err != nil {
		return err
	}

	if !claimed {
		// node already claimed by other consumer
		return nil
	}

	log.Printf("claimed run %q node %q\n", event.RunID, event.NodeSpec.Name)

	nodeResult := &adagio.Node_Result{}

	switch event.Type {
	case adagio.Event_NODE_READY:
		var result *adagio.Result
		if result, err = runtime.Run(node); err == nil {
			nodeResult = &adagio.Node_Result{
				Conclusion: adagio.Node_Result_Conclusion(result.Conclusion),
				Metadata:   result.Metadata,
				Output:     result.Output,
			}
		}

	case adagio.Event_NODE_ORPHANED:
		err = errors.New("node was orphaned")
	}

	if err != nil {
		nodeResult.Conclusion = adagio.Node_Result_ERROR
		nodeResult.Output = []byte(err.Error())
	}

	if err := p.repo.FinishNode(event.RunID, event.NodeSpec.Name, nodeResult, claim); err != nil {
		return err
	}

	return nil
}
