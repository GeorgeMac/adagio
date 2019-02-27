package worker

import (
	"context"
	"log"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/repository"
)

type Handler interface {
	Run(*adagio.Node) error
}

type HandlerFunc func(*adagio.Node) error

func (fn HandlerFunc) Run(n *adagio.Node) error { return fn(n) }

type Pool struct {
	repo    repository.Repository
	handler Handler

	size int
}

func NewPool(repo repository.Repository, handler Handler) *Pool {
	return &Pool{
		repo:    repo,
		handler: handler,
		size:    5,
	}
}

func (p *Pool) Run(ctxt context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events := make(chan repository.Event, 10)
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

					if err := p.handler.Run(event.Node); err != nil {
						log.Println(err)
						// TODO implement retry behavior
						continue
					}

					if err := p.repo.FinishNode(event.Run, event.Node); err != nil {
						log.Println(err)
					}

					log.Println("finished", event.Node)
				case <-ctxt.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
}
