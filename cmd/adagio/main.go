package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

func printUsage() {
	fmt.Println("usage: adagio <command>")
	fmt.Println("              runs")
	fmt.Println("              help")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "runs":
		runs()
	case "help":
		printUsage()
	default:
		printUsage()
		os.Exit(1)
	}
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
