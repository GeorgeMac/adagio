package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

func main() {
	var (
		fs   = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		host = fs.String("host", "http://localhost:7890", "host address of adagio control plane")
		_    = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Println("Usage: adagio <COMMAND> [OPTIONS]\n")
		fmt.Println("Commands:")
		fmt.Println("\truns - manage adagio runs")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}

	var (
		ctxt   = context.Background()
		client = controlplane.NewControlPlaneProtobufClient(*host, &http.Client{})
	)

	switch fs.Arg(0) {
	case "runs":
		runs(ctxt, client, fs.Args())
	default:
		fs.Usage()
		os.Exit(2)
	}
}
