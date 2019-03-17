package worker

import (
	"context"
	"log"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

type Repository interface {
	ClaimNode(*adagio.Run, *adagio.Node) (bool, error)
	FinishNode(*adagio.Run, *adagio.Node) error
	Subscribe(events chan<- adagio.Event, states ...adagio.NodeState) error
}

type Runtime interface {
	Run(*adagio.Node) error
}

type RuntimeFunc func(*adagio.Node) error

func (fn RuntimeFunc) Run(n *adagio.Node) error { return fn(n) }

type Pool struct {
	repo    Repository
	runtime Runtime

	size int
}

func NewPool(repo Repository, runtime Runtime) *Pool {
	return &Pool{
		repo:    repo,
		runtime: runtime,
		size:    5,
	}
}

func (p *Pool) Run(ctxt context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.size; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			events := make(chan adagio.Event, 10)
			p.repo.Subscribe(events, adagio.ReadyState)

			for {
				select {
				case event := <-events:
					claimed, err := p.repo.ClaimNode(event.Run, event.Node)
					if err != nil {
						log.Println(err)
					}

					if !claimed {
						log.Println("node already claimed")
						continue
					}

					if err := p.runtime.Run(event.Node); err != nil {
						log.Println(err)
						// TODO implement retry behavior
						continue
					}

					if err := p.repo.FinishNode(event.Run, event.Node); err != nil {
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
