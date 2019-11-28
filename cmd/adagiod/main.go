package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/georgemac/adagio/pkg/agent"
	"github.com/georgemac/adagio/pkg/etcd"
	"github.com/georgemac/adagio/pkg/memory"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/georgemac/adagio/pkg/runtimes/debug"
	"github.com/georgemac/adagio/pkg/runtimes/exec"
	controlservice "github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftoml"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

type Repository interface {
	controlservice.Repository
	agent.Repository
}

func main() {
	var (
		fs        = flag.NewFlagSet("adagiod", flag.ExitOnError)
		backend   = fs.String("backend-type", "memory", `backend repository type ("memory"|"etcd")`)
		etcdAddrs = fs.String("etcd-addresses", "http://127.0.0.1:2379", "list of etcd node addresses")
		_         = fs.String("config", "", "location of config toml file")

		ctxt, cancel     = context.WithCancel(context.Background())
		runAPI, runAgent = true, true

		repo Repository
		wg   sync.WaitGroup
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagiod <api|agent> [OPTIONS]\n\n")
		fmt.Print("The adagio workflow agent and control plane API\n\n")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(fftoml.Parser),
		ff.WithEnvVarPrefix("ADAGIOD"))

	switch *backend {
	case "memory":
		repo = memory.New()
	case "etcd":
		endpoints := strings.Split(*etcdAddrs, ",")
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: 3 * time.Second,
		})
		if err != nil {
			log.Fatal(err)
		}

		repo = etcd.New(cli.KV, cli.Watcher, cli.Lease)
	default:
		fmt.Printf("unexpected backend repository type %q expected one of [memory|etcd]\n", *backend)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop

		cancel()
	}()

	if len(fs.Args()) > 0 {
		switch fs.Arg(0) {
		case "agent":
			runAPI = false
		case "api":
			runAgent = false
		}
	}

	if runAPI {
		wg.Add(1)
		go func() {
			defer wg.Done()

			startAPI(ctxt, repo)
		}()
	}

	if runAgent {
		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Printf("Agent accepting work from %q backend\n", *backend)

			startAgents(ctxt, repo)
		}()
	}

	wg.Wait()
}

func startAPI(ctxt context.Context, repo controlservice.Repository) {
	var (
		service       = controlservice.New(repo)
		addr          = ":7890"
		grpcServer    = grpc.NewServer()
		listener, err = net.Listen("tcp", addr)
	)

	if err != nil {
		log.Fatal(err)
	}

	controlplane.RegisterControlPlaneServer(grpcServer, service)

	log.Printf("Control plane listening on %q\n", addr)

	go func() {
		<-ctxt.Done()

		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Println(err)
	}
}

func startAgents(ctxt context.Context, repo agent.Repository) {
	runtimes := agent.RuntimeMap{}
	runtimes.Register(exec.Runtime())
	runtimes.Register(debug.Runtime())

	agent.NewPool(repo, runtimes, agent.WithAgentCount(5)).Run(ctxt)
}
