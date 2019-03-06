package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/georgemac/adagio/internal/controlplaneservice"
	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/memory"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
)

func main() {
	var (
		repo    = memory.NewRepository()
		handler = worker.HandlerFunc(func(node *adagio.Node) error {
			fmt.Printf("got node %s\n", node)
			time.Sleep(5 * time.Second)
			fmt.Printf("finished with node %s\n", node)
			return nil
		})
		pool    = worker.NewPool(repo, handler)
		service = controlplaneservice.New(repo)
		mux     = controlplane.NewControlPlaneServer(service, nil)
		server  = &http.Server{
			Addr:    ":7890",
			Handler: mux,
		}
		ctxt, cancel = context.WithCancel(context.Background())
	)

	var (
		stop     = make(chan os.Signal, 1)
		finished = make(chan struct{})
	)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		defer close(finished)

		<-stop
		cancel()

		server.Shutdown(context.Background())
	}()

	// run worker pool
	go pool.Run(ctxt)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-finished
}
