// +build etcd

package etcd

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/repository"
	"go.etcd.io/etcd/clientv3"
)

func Test_Run_RepositoryTestHarness(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_, err := cli.KV.Delete(context.Background(), "adagio-test/", clientv3.WithPrefix())
		if err != nil {
			log.Fatal(err)
		}
	}()

	var (
		repo     = New(cli.KV, cli.Watcher, cli.Lease, WithNamespace("adagio-test/"))
		orphaner = repository.OrphanFunc(func(c *adagio.Claim) {
			repo.cancelLease(c.Id)
		})
	)

	repository.TestHarness(t, func(now func() time.Time) (repository.Repository, repository.Orphaner) {
		repo.now = now

		return repo, orphaner
	})
}
