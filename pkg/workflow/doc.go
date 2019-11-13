//	package main
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
//			a       = MustNode(builder.Node("a", success))
//			b       = MustNode(builder.Node("b", success))
//			c       = MustNode(builder.Node("c", success))
//			d       = MustNode(builder.Node("d", potentialPanic, workflow.WithRetry(adagio.OnError, 3)))
//			e       = MustNode(builder.Node("e", success))
//			f       = MustNode(builder.Node("f", success))
//			g       = MustNode(builder.Node("g", runLS))
//		)
//
//		c.DependsOn(a)
//		d.DependsOn(a, b)
//		f.DependsOn(b)
//		e.DependsOn(c, d)
//		g.DependsOn(e, f)
//
//		return builder.Start(ctxt, client)
//	}
//
package workflow
