package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/etcd"
	"github.com/georgemac/adagio/pkg/memory"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/georgemac/adagio/pkg/runtimes/exec"
	controlservice "github.com/georgemac/adagio/pkg/service/controlplane"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/fftoml"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

type Repository interface {
	controlservice.Repository
	worker.Repository
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

			api(ctxt, repo)
		}()
	}

	if runAgent {
		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Printf("Agent accepting work from %q backend\n", *backend)

			agent(ctxt, repo)
		}()
	}

	wg.Wait()
}

func api(ctxt context.Context, repo controlservice.Repository) {
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

func agent(ctxt context.Context, repo worker.Repository) {
	var (
		runtimes = map[string]worker.Runtime{
			"echo": worker.RuntimeFunc(func(node *adagio.Node) (*adagio.Result, error) {
				return &adagio.Result{
					Conclusion: adagio.Result_SUCCESS,
					Output:     []byte(node.Spec.Name),
				}, nil
			}),
			"flakey": worker.RuntimeFunc(func(node *adagio.Node) (*adagio.Result, error) {
				if rand.Intn(2) > 0 {
					return &adagio.Result{Conclusion: adagio.Result_FAIL}, nil
				}

				return &adagio.Result{Conclusion: adagio.Result_SUCCESS}, nil
			}),
			"fail": worker.RuntimeFunc(func(node *adagio.Node) (*adagio.Result, error) {
				return &adagio.Result{Conclusion: adagio.Result_FAIL}, nil
			}),
			"error": worker.RuntimeFunc(func(node *adagio.Node) (*adagio.Result, error) {
				return &adagio.Result{}, errors.New("something went wrong")
			}),
			"panic": worker.RuntimeFunc(func(node *adagio.Node) (*adagio.Result, error) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				if x := r.Intn(10); x > 4 {
					fmt.Println("got:", x)
					panic("uh oh")
				}

				return &adagio.Result{Conclusion: adagio.Result_SUCCESS}, nil
			}),
			"exec": exec.New(),
		}
	)

	worker.NewPool(repo, runtimes, worker.WithWorkerCount(5)).Run(ctxt)
}
