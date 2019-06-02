package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/georgemac/adagio/pkg/adagio"
)

var ErrRuntimeDoesNotExist = errors.New("runtime does not exist")

type Repository interface {
	ClaimNode(runID, name string) (*adagio.Node, bool, error)
	FinishNode(runID, name string) error
	Subscribe(events chan<- *adagio.Event, states ...adagio.Node_Status) error
}

type Runtime interface {
	Run(*adagio.Node) (*adagio.Result, error)
}

type RuntimeFunc func(*adagio.Node) (*adagio.Result, error)

func (fn RuntimeFunc) Run(n *adagio.Node) (*adagio.Result, error) { return fn(n) }

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
	node, claimed, err := p.repo.ClaimNode(event.RunID, event.NodeName)
	if err != nil {
		return err
	}

	if !claimed {
		return errors.New("node already claimed")
	}

	runtime, ok := p.runtimes[node.Spec.Runtime]
	if !ok {
		return ErrRuntimeDoesNotExist
	}

	resp, err := runtime.Run(node)
	if err != nil {
		// TODO implement retry behavior
		return err
	}

	fmt.Println(string(resp.Output))

	if err := p.repo.FinishNode(event.RunID, event.NodeName); err != nil {
		return err
	}

	return nil
}
