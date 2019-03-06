package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

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
		start(client)
	case "ls":
		resp, err := client.List(context.Background(), &controlplane.ListRequest{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Fprintln(w, "Runs\tCreated At\t")
		for _, run := range resp.Runs {
			fmt.Fprintf(w, "%s\t%s\t\n", run.Id, run.CreatedAt)
		}
		w.Flush()
	default:
		printRunsUsage()
		os.Exit(1)
	}
}

func printStartUsage() {
	fmt.Println("usage: adagio start [file]")
	fmt.Println("                    <stdin>")
}

func start(client controlplane.ControlPlane) {
	var (
		graph = &controlplane.Graph{}
		input io.Reader
	)

	if len(os.Args) < 4 {
		input = os.Stdin
	} else {
		fi, err := os.Open(os.Args[3])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer fi.Close()

		input = fi
	}

	if err := json.NewDecoder(input).Decode(graph); err != nil {
		fmt.Println(err)
		os.Exit(1)

	}

	run, err := client.Start(context.Background(), graph)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Run started %q\n", run.Id)
}
