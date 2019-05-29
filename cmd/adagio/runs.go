package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

func runs(ctxt context.Context, client controlplane.ControlPlane, args []string) {
	var (
		fs = flag.NewFlagSet(args[0], flag.ExitOnError)
		_  = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Println("Usage: adagio runs <COMMAND> [OPTIONS]\n")
		fmt.Println("Commands:")
		fmt.Println("\tstart - starts a new run from the provided graph spec")
		fmt.Println("\tls    - list current and previous runs")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}

	switch fs.Arg(0) {
	case "start":
		start(ctxt, client, fs.Args())
	case "ls":
		list(ctxt, client)
	default:
		fs.Usage()
		os.Exit(2)
	}
}

func start(ctxt context.Context, client controlplane.ControlPlane, args []string) {
	var (
		fs = flag.NewFlagSet(args[0], flag.ExitOnError)
		_  = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Println("Usage: adagio runs start [OPTIONS]\n")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	var (
		req = &controlplane.StartRequest{
			Spec: &adagio.GraphSpec{},
		}
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

	if err := json.NewDecoder(input).Decode(req.Spec); err != nil {
		fmt.Println(err)
		os.Exit(1)

	}

	resp, err := client.Start(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Run started %q\n", resp.Run.Id)
}

func list(ctxt context.Context, client controlplane.ControlPlane) {
	resp, err := client.List(ctxt, &controlplane.ListRequest{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)

	fmt.Fprintln(w, "ID\tCreated At\t")
	for _, run := range resp.Runs {
		fmt.Fprintf(w, "%s\t%s\t\n", run.Id, run.CreatedAt)
	}

	w.Flush()
}
