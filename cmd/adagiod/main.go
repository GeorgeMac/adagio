package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/etcd"
	"github.com/georgemac/adagio/pkg/memory"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	controlservice "github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
	"go.etcd.io/etcd/clientv3"
)

type Repository interface {
	controlservice.Repository
	worker.Repository
}

func main() {
	var (
		repository       = flag.String("repo", "memory", "repository type [memory|etcd]")
		addrs            = flag.String("etcd-addresses", "http://127.0.0.1:2379", "list of etcd node addresses")
		repo             Repository
		ctxt, cancel     = context.WithCancel(context.Background())
		wg               sync.WaitGroup
		runAPI, runAgent = true, true
	)

	flag.Parse()

	switch *repository {
	case "memory":
		repo = memory.New()
	case "etcd":
		cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(*addrs, ",")})
		if err != nil {
			log.Fatal(err)
		}

		repo = etcd.New(cli.KV, cli.Watcher)
	default:
		fmt.Printf("unexpected repository type %q expected one of [memory|etcd]\n", *repository)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop

		cancel()
	}()

	if len(flag.Args()) > 0 {
		switch flag.Arg(0) {
		case "agent":
			runAPI = false
		case "api":
			runAgent = false
		}
	}

	if runAPI {
		wg.Add(1)
		go func() {
			defer wg.Done()

			api(ctxt, repo)
		}()
	}

	if runAgent {
		wg.Add(1)
		go func() {
			defer wg.Done()

			agent(ctxt, repo)
		}()
	}

	wg.Wait()
}

func api(ctxt context.Context, repo controlservice.Repository) {
	var (
		service = controlservice.New(repo)
		mux     = controlplane.NewControlPlaneServer(service, nil)
		server  = &http.Server{
			Addr:    ":7890",
			Handler: mux,
		}
	)

	go func() {
		<-ctxt.Done()

		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func agent(ctxt context.Context, repo worker.Repository) {
	var (
		runtimes = map[string]worker.Runtime{
			"echo": worker.RuntimeFunc(func(node *adagio.Node) error {
				fmt.Printf("got node %s\n", node)
				return nil
			}),
		}
	)

	worker.NewPool(repo, runtimes).Run(ctxt)
}
