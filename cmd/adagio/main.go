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

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/api"
	"github.com/georgemac/adagio/pkg/repository/memory"
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
		pool   = worker.NewPool(repo, handler)
		api    = api.NewServer(repo)
		server = &http.Server{
			Addr:    ":7890",
			Handler: api,
		}
		ctxt, cancel = context.WithCancel(context.Background())
	)

	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Println("usage: adagio <command>")
		fmt.Println("              serve")
		os.Exit(1)
	}

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
