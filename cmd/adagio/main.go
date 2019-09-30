package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"google.golang.org/grpc"
)

func main() {
	var (
		fs   = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		host = fs.String("host", "localhost:7890", "host address of adagio control plane")
		_    = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagio <COMMAND> [OPTIONS]\n\n")
		fmt.Println("Commands:")
		fmt.Println("\truns  - manage adagio runs")
		fmt.Println("\tstats - view adagio statistics")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])

	if fs.NArg() < 1 {
		exit(fs.Usage, 2)
	}

	conn, err := grpc.Dial(*host, grpc.WithInsecure())
	exitIfError(err)

	defer conn.Close()

	switch fs.Arg(0) {
	case "runs":
		runs(context.Background(), controlplane.NewControlPlaneClient(conn), fs.Args())
	case "stats":
		stats(context.Background(), controlplane.NewControlPlaneClient(conn), fs.Args())
	default:
		exit(fs.Usage, 2)
	}
}

func stats(ctxt context.Context, client controlplane.ControlPlaneClient, args []string) {
	var (
		fs = flag.NewFlagSet(args[0], flag.ExitOnError)
		_  = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagio stats [OPTIONS]\n\n")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	resp, err := client.Stats(ctxt, &controlplane.StatsRequest{})
	exitIfError(err)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	stats := resp.Stats

	fmt.Fprintf(w, "runs\t%d\t\n", stats.RunCount)
	fmt.Fprintf(w, "nodes waiting\t%d\t\n", stats.NodeCounts.WaitingCount)
	fmt.Fprintf(w, "nodes ready\t%d\t\n", stats.NodeCounts.ReadyCount)
	fmt.Fprintf(w, "nodes running\t%d\t\n", stats.NodeCounts.RunningCount)
	fmt.Fprintf(w, "nodes completed\t%d\t\n", stats.NodeCounts.CompletedCount)

	w.Flush()
}

func exitIfError(err error) {
	if err == nil {
		return
	}

	exit(err, 1)
}

func exit(v interface{}, code int) {
	if fn, ok := v.(func()); ok {
		fn()
		os.Exit(code)
	}

	fmt.Println(v)
	os.Exit(code)
}
