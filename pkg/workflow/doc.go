// Package workflow
//
// The workflow package contains a helpful builder type which provides fluent style functions to construct a graph
// and to execute the graph on a provided control plane client.
//
//  package main
//
//	import (
//		"context"
//
//		"github.com/georgemac/adagio/pkg/adagio"
//		"github.com/georgemac/adagio/pkg/rpc/controlplane"
//		"github.com/georgemac/adagio/pkg/runtimes/debug"
//		"github.com/georgemac/adagio/pkg/runtimes/exec"
//		"github.com/georgemac/adagio/pkg/workflow"
//	)
//
//  // start creates, configures and runs the following graph workflow:
//  // (a) ---> (c)----
//  //   \             \
//  //    ------v       v
//  //         (d) --> (e) --> (g)
//  //    ------^               ^
//  //   /                     /
//  // (b) --> (f) ------------
//  // Node (d) is configured to panic 50% of the time with up to 3 attempts on this kind of system error
//  // Node (g) is configured to ls / on the running agent
//	func start(ctxt context.Context, client controlplane.ControlPlaneClient) (*adagio.Run, error) {
//		var (
//			success        = debug.NewCall(adagio.Result_SUCCESS)
//			potentialPanic = debug.NewCall(adagio.Result_SUCCESS,
//				debug.With(debug.Chance(0.5, debug.Panic)))
//
//			runLS   = exec.NewCall("ls", "/")
//
//			builder = workflow.NewBuilder()
//			a       = builder.Node("a", success)
//			b       = builder.Node("b", success)
//			c       = builder.Node("c", success)
//			d       = builder.Node("d", potentialPanic, workflow.WithRetry(adagio.OnError, 3))
//			e       = builder.Node("e", success)
//			f       = builder.Node("f", success)
//			g       = builder.Node("g", runLS)
//		)
//
//		c.DependsOn(a)
//		d.DependsOn(a)
//		d.DependsOn(b)
//		f.DependsOn(b)
//		e.DependsOn(d)
//		e.DependsOn(c)
//		g.DependsOn(e)
//		g.DependsOn(f)
//
//		return builder.Start(ctxt, client)
//	}
//
package workflow
