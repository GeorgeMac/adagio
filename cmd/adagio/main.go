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

func printUsage() {
	fmt.Println("usage: adagio <command>")
	fmt.Println("              serve")
	fmt.Println("              runs")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serve()
	case "runs":
		runs()
	default:
		printUsage()
		os.Exit(1)
	}
}

func serve() {
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

func printRunsUsage() {
	fmt.Println("usage: adagio runs <command>")
	fmt.Println("                   start")
	fmt.Println("                   ls")
}

func runs() {
	if len(os.Args) < 3 {
		printRunsUsage()
		os.Exit(1)
	}

	client := controlplane.NewControlPlaneProtobufClient("http://localhost:7890", &http.Client{})

	switch os.Args[2] {
	case "start":
		graph := &controlplane.Graph{
			Nodes: []*controlplane.Node{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
				{Name: "d"},
				{Name: "e"},
				{Name: "f"},
				{Name: "g"},
			},
			Edges: []*controlplane.Edge{
				{Source: "a", Destination: "c"},
				{Source: "a", Destination: "d"},
				{Source: "b", Destination: "d"},
				{Source: "b", Destination: "f"},
				{Source: "c", Destination: "e"},
				{Source: "d", Destination: "e"},
				{Source: "e", Destination: "g"},
				{Source: "f", Destination: "g"},
			},
		}
		run, err := client.Start(context.Background(), graph)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Run started %q\n", run.Id)
	case "ls":
		resp, err := client.List(context.Background(), &controlplane.ListRequest{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Runs")
		for _, run := range resp.Runs {
			fmt.Println(run.Id)
		}
	default:
		printRunsUsage()
		os.Exit(1)
	}
}
