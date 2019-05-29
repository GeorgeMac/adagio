// +build etcd

package etcd

import (
	"context"
	"log"
	"testing"
	"time"

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

	repo := New(cli.KV, cli.Watcher, WithNamespace("adagio-test/"))

	repository.TestHarness(t, func(now func() time.Time) repository.Repository {
		repo.now = now

		return repo
	})
}
