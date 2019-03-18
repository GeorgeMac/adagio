package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/memory"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	controlservice "github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
)

func main() {
	var (
		repo             = memory.New()
		ctxt, cancel     = context.WithCancel(context.Background())
		wg               sync.WaitGroup
		runAPI, runAgent = true, true
	)

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop

		cancel()
	}()

	if len(os.Args) > 1 {
		switch os.Args[1] {
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
	worker.NewPool(repo, worker.RuntimeFunc(func(node *adagio.Node) error {
		fmt.Printf("got node %s\n", node)
		time.Sleep(5 * time.Second)
		fmt.Printf("finished with node %s\n", node)
		return nil
	})).Run(ctxt)
}
