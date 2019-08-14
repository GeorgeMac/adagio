package memory

import (
	"testing"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/repository"
)

func Test_Run_RepositoryTestHarness(t *testing.T) {
	repo := New()

	repository.TestHarness(t, func(now func() time.Time) (repository.Repository, repository.Orphaner) {
		repo.now = now

		return repo, repository.OrphanFunc(func(claim *adagio.Claim) {
			// in-memory repo doesn't really orphan nodes as there
			// are no potential network related issues
			// however, this emulates the failure scenario in-order to
			// satisfy the test constraints
			c, ok := repo.claims[claim.Id]
			if !ok {
				t.Fatal("node not found for claim")
			}

			// set node status to none to signify node has been orphaned
			c.node.Status = adagio.Node_NONE

			// notify listens of orphan
			for _, ch := range repo.listeners[adagio.Event_NODE_ORPHANED] {
				select {
				case ch <- &adagio.Event{RunID: c.run.Id, NodeSpec: c.node.Spec, Type: adagio.Event_NODE_ORPHANED}:
					// attempt to send
				default:
				}
			}
		})
	})
}
