package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
		fmt.Println("\truns - manage adagio runs")
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
