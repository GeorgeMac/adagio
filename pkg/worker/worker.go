package worker

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

var ErrRuntimeDoesNotExist = errors.New("runtime does not exist")

type Repository interface {
	ClaimNode(run *adagio.Run, name string) (*adagio.Node, bool, error)
	FinishNode(run *adagio.Run, name string) error
	Subscribe(events chan<- *adagio.Event, states ...adagio.Node_State) error
}

type Runtime interface {
	Run(*adagio.Node) error
}

type RuntimeFunc func(*adagio.Node) error

func (fn RuntimeFunc) Run(n *adagio.Node) error { return fn(n) }

type Pool struct {
	repo     Repository
	runtimes map[string]Runtime

	size int
}

func NewPool(repo Repository, runtimes map[string]Runtime) *Pool {
	return &Pool{
		repo:     repo,
		runtimes: runtimes,
		size:     5,
	}
}

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
					node, claimed, err := p.repo.ClaimNode(event.Run, event.Node.Spec.Name)
					if err != nil {
						log.Println(err)
						continue
					}

					if !claimed {
						log.Println("node already claimed")
						continue
					}

					runtime, ok := p.runtimes[event.Node.Spec.Runtime]
					if !ok {
						log.Println(ErrRuntimeDoesNotExist)
						continue
					}

					if err := runtime.Run(node); err != nil {
						log.Println(err)
						// TODO implement retry behavior
						continue
					}

					if err := p.repo.FinishNode(event.Run, event.Node.Spec.Name); err != nil {
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
