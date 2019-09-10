package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	adagiodAddr = flag.String("adagiod-addr", "localhost:7890", "gRPC server endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mux  = runtime.NewServeMux()
		opts = []grpc.DialOption{grpc.WithInsecure()}
	)

	if err := controlplane.RegisterControlPlaneHandlerFromEndpoint(ctx, mux, *adagiodAddr, opts); err != nil {
		return err
	}

	return http.ListenAndServe(":7891", mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
