package worker

import (
	"errors"

	"github.com/georgemac/adagio/pkg/adagio"
)

// NodeRuntimeFunc is a type used to simplify Runtime
// creation to a single anonymous function
type NodeRuntime struct {
	name string
	call Call
}

// Name return the name of the NodeRuntimeFunc
func (n NodeRuntime) Name() string { return n.name }

// BlankCall returns the underlying call
func (n NodeRuntime) BlankCall() Call { return n.call }

// NodeRuntimeFunc takes a name and anonymous runtime function
// and converts it into a Runtime
// It's purpose is to aid in simply runtime definition
func NodeRuntimeFunc(name string, runner NodeRunner) Runtime {
	return NodeRuntime{
		name: name,
		call: CallFunc(runner),
	}
}

// NodeRunner is a function which executes the instructions
// defined within a node and produces a result
type NodeRunner func(*adagio.Node) (*adagio.Result, error)

// CallFunc converts a function which produces an *adagio.Result from
// an *adagio.Node into a Call implementation
func CallFunc(fn NodeRunner) Call {
	return &NodeCaller{run: fn}
}

// NodeCaller implements Call and delegates to an underyling
// anonymous run function
// It passes nodes provided on calls to Parse to the anonymous
// function on calls to Run and returns the result
type NodeCaller struct {
	n   *adagio.Node
	run NodeRunner
}

// Parse captures the Node
func (r *NodeCaller) Parse(n *adagio.Node) error {
	r.n = n
	return nil
}

// Run delegates the captured Node onto the decorate function
func (r *NodeCaller) Run() (*adagio.Result, error) {
	if r.n == nil {
		return nil, errors.New("node not provided on parse")
	}

	return r.run(r.n)
}
