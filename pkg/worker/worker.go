package worker

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

// ErrRuntimeDoesNotExist is returned when a node is claimed with an
// unkown runtime type
var ErrRuntimeDoesNotExist = errors.New("runtime does not exist")

// Repository is the minimal interface for a backing repository which can
// notify of node related events, issue node claims and finalize the result
// of executing a node
type Repository interface {
	ClaimNode(runID, name string) (*adagio.Node, bool, error)
	FinishNode(runID, name string, result *adagio.Result) error
	Subscribe(events chan<- *adagio.Event, states ...adagio.Node_Status) error
}

// Runtime is a type which can execute a node and produce a result
type Runtime interface {
	Run(*adagio.Node) (*adagio.Result, error)
}

// RuntimeFunc is a function which can be used as a Runtime
type RuntimeFunc func(*adagio.Node) (*adagio.Result, error)

// Run delegates to the wrapped RuntimeFunc
func (fn RuntimeFunc) Run(n *adagio.Node) (*adagio.Result, error) { return fn(n) }

// Pool spins up a number of worker goroutines which subscribe to nodes
// transitioning into the ready state and then attempts to claim and
// process them
type Pool struct {
	repo     Repository
	runtimes map[string]Runtime

	size int
}

// NewPool constructs and configures a new node pool for execution
func NewPool(repo Repository, runtimes map[string]Runtime, opts ...Option) *Pool {
	pool := &Pool{
		repo:     repo,
		runtimes: runtimes,
		size:     1,
	}

	Options(opts).Apply(pool)

	return pool
}

// Run begins the configured number of workers and responds to cancelation
// of the supplied context
func (p *Pool) Run(ctxt context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.size; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			events := make(chan *adagio.Event, 10)
			p.repo.Subscribe(events, adagio.Node_READY)

			for {
				select {
				case event := <-events:
					if err := p.handleEvent(event); err != nil {
						log.Println(err)
					}

				case <-ctxt.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
}

func (p *Pool) handleEvent(event *adagio.Event) error {
	runtime, ok := p.runtimes[event.NodeSpec.Runtime]
	if !ok {
		return ErrRuntimeDoesNotExist
	}

	node, claimed, err := p.repo.ClaimNode(event.RunID, event.NodeSpec.Name)
	if err != nil {
		return err
	}

	if !claimed {
		return errors.New("node already claimed")
	}

	result, err := runtime.Run(node)
	if err != nil {
		// TODO implement retry behavior
		return err
	}

	if err := p.repo.FinishNode(event.RunID, event.NodeSpec.Name, result); err != nil {
		return err
	}

	return nil
}
