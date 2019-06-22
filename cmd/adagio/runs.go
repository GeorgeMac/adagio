package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/printing"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

func runs(ctxt context.Context, client controlplane.ControlPlane, args []string) {
	var (
		fs = flag.NewFlagSet(args[0], flag.ExitOnError)
		_  = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagio runs <COMMAND> [OPTIONS]\n\n")
		fmt.Println("Commands:")
		fmt.Println("\tstart - starts a new run from the provided graph spec")
		fmt.Println("\tls    - list current and previous runs")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	if fs.NArg() < 1 {
		exit(fs.Usage, 2)
	}

	switch fs.Arg(0) {
	case "start":
		start(ctxt, client, fs.Args()...)
	case "inspect":
		inspect(ctxt, client, fs.Args()...)
	case "ls":
		list(ctxt, client)
	default:
		exit(fs.Usage, 2)
	}
}

func start(ctxt context.Context, client controlplane.ControlPlane, args ...string) {
	var (
		fs = flag.NewFlagSet(args[0], flag.ExitOnError)
		_  = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagio runs start [OPTIONS]\n\n")
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

	exitIfError(json.NewDecoder(input).Decode(req.Spec))

	resp, err := client.Start(context.Background(), req)
	exitIfError(err)

	fmt.Printf("Run started %q\n", resp.Run.Id)
}

func inspect(ctxt context.Context, client controlplane.ControlPlane, args ...string) {
	var (
		fs      = flag.NewFlagSet(args[0], flag.ExitOnError)
		printer = fs.String("printer", "pretty", `printer to use ("pretty"|"spew"|"dot") (default "pretty")`)
		format  = fs.String("format", "", `text/template string (e.g. "{{ .ID }} {{ .CreatedAt }}")`)
		_       = fs.Bool("help", false, "print usage")
	)

	fs.Usage = func() {
		fmt.Println()
		fmt.Print("Usage: adagio runs inspect [OPTIONS] <run_id>\n\n")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args[1:])

	if fs.NArg() < 1 {
		exit(fs.Usage, 2)
	}

	resp, err := client.Inspect(ctxt, &controlplane.InspectRequest{
		Id: fs.Arg(0),
	})
	exitIfError(err)

	if *format != "" {
		var (
			fnMap = template.FuncMap{
				"format": time.Time.Format,
				"using": func(fmt string) string {
					switch strings.ToLower(fmt) {
					case "ansic":
						return time.ANSIC
					case "unixdate":
						return time.UnixDate
					case "rubydate":
						return time.RubyDate
					case "rfc822":
						return time.RFC822
					case "rfc822z":
						return time.RFC822Z
					case "rfc850":
						return time.RFC850
					case "rfc1123":
						return time.RFC1123
					case "rfc1123z":
						return time.RFC1123Z
					case "rfc3339":
						return time.RFC3339
					case "rfc3339nano":
						return time.RFC3339Nano
					case "kitchen":
						return time.Kitchen
					case "stamp":
						return time.Stamp
					case "stampmilli":
						return time.StampMilli
					case "stampmicro":
						return time.StampMicro
					case "stampnano":
						return time.StampNano
					}

					return ""
				},
			}
			tmpl = template.Must(template.
				New("template").
				Funcs(fnMap).
				Parse(*format))
			run, err = printing.PBRunToRun(resp.Run)
		)

		exitIfError(err)

		exitIfError(tmpl.Execute(os.Stdout, run))

		fmt.Println()

		return
	}

	var formatter fmt.Formatter
	switch *printer {
	case "pretty":
		formatter, err = printing.Pretty(resp.Run)
	case "spew":
		formatter, err = printing.Spew(resp.Run)
	case "dot":
		formatter, err = printing.Dot(resp.Run)
	default:
		fmt.Println(*printer, "printer not recognized")
		os.Exit(1)
	}

	exitIfError(err)

	fmt.Printf("%# v\n", formatter)
}

func list(ctxt context.Context, client controlplane.ControlPlane) {
	resp, err := client.List(ctxt, &controlplane.ListRequest{})
	exitIfError(err)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)

	fmt.Fprintln(w, "ID\tCreated At\t")
	for _, run := range resp.Runs {
		fmt.Fprintf(w, "%s\t%s\t\n", run.Id, run.CreatedAt)
	}

	w.Flush()
}
