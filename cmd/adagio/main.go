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
		fmt.Print("Usage: adagio <COMMAND> [OPTIONS]\n\n")
		fmt.Println("Commands:")
		fmt.Println("\truns - manage adagio runs")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])

	if fs.NArg() < 1 {
		exit(fs.Usage, 2)
	}

	var (
		ctxt   = context.Background()
		client = controlplane.NewControlPlaneProtobufClient(*host, &http.Client{})
	)

	switch fs.Arg(0) {
	case "runs":
		runs(ctxt, client, fs.Args())
	default:
		exit(fs.Usage, 2)
	}
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
