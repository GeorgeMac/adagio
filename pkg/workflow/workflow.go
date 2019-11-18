package workflow

import (
	"context"
	"fmt"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

// Function is a type which can be composed into a workflow
// It represents itself as a node specifications and can be
// invoked in the graph as a node
type Function interface {
	NewSpec(name string) (*adagio.Node_Spec, error)
}

// InputMappableFunction is a Function which allows for
// its arguments to be mapped from node inputs
type InputMappableFunction interface {
	Function
	SetArgumentFromInput(argument, input string) error
}

// NodeOption is a functional option for a node spec
type NodeOption func(*adagio.Node_Spec)

// NodeOptions is a slice of NodeOption types
type NodeOptions []NodeOption

// Apply calls each option in turn on the provided Node_Spec
func (o NodeOptions) Apply(spec *adagio.Node_Spec) {
	for _, opt := range o {
		opt(spec)
	}
}

// WithRetry configures a retry for a specified condition up to
// maxAttempts times on a Node_Spec
func WithRetry(condition adagio.RetryCondition, maxAttempts int32) NodeOption {
	return func(spec *adagio.Node_Spec) {
		if spec.Retry == nil {
			spec.Retry = map[string]*adagio.Node_Spec_Retry{}
		}

		spec.Retry[string(condition)] = &adagio.Node_Spec_Retry{MaxAttempts: maxAttempts}
	}
}

// Builder is a type used to compose calls to start runs on a client
// It can be used to convert runtime calls into workflow nodes
// configure connections between nodes and then invoke the
// built graph spec onto a client which produces a new Run
type Builder struct {
	err   error
	nodes []Node
	edges []*adagio.Edge
}

// NewBuilder creates and configures a new Builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Node creates a node from a provided name and SpecBuilder type
// it also applies any provided options
func (b *Builder) Node(name string, fn Function, opts ...NodeOption) (n Node) {
	n = Node{b, name, fn, NodeOptions(opts)}

	b.nodes = append(b.nodes, n)

	return
}

// Build constructs a graph specification from the builders state
func (b *Builder) Build() (*adagio.GraphSpec, error) {
	if b.err != nil {
		return nil, b.err
	}

	spec := &adagio.GraphSpec{
		Nodes: make([]*adagio.Node_Spec, 0, len(b.nodes)),
		Edges: b.edges,
	}

	for _, node := range b.nodes {
		nSpec, err := node.fn.NewSpec(node.name)
		if err != nil {
			return nil, err
		}

		node.opts.Apply(nSpec)
		spec.Nodes = append(spec.Nodes, nSpec)
	}

	return spec, nil
}

// Start invokes the built graph specification onto the provided client
// It returns the run responded by the controlplane API
func (b *Builder) Start(ctx context.Context, client controlplane.ControlPlaneClient) (*adagio.Run, error) {
	spec, err := b.Build()
	if err != nil {
		return nil, err
	}

	resp, err := client.Start(ctx, &controlplane.StartRequest{Spec: spec})
	if err != nil {
		return nil, err
	}

	return resp.Run, nil
}

// Node is a builder wrapper type which can be used to further
// create connections between nodes in the originating builder
type Node struct {
	builder *Builder
	name    string
	fn      Function
	opts    NodeOptions
}

// DependencyOption is a function which manipulates a dependency
// between two nodes
type DependencyOption func(from, on Node)

// MapOutputTo maps the output of the dependency onto
// the argument name of the callee
func MapOutputTo(argument string) DependencyOption {
	return func(from, on Node) {
		if fn, ok := from.fn.(InputMappableFunction); ok {
			fn.SetArgumentFromInput(argument, on.name)
			return
		}

		from.builder.err = fmt.Errorf("argument %q does not support input mapping", from.name)
	}
}

// DependsOn creates a connection from the provided nodes (sources)
// to the callee node (destination) on the original builder
func (n Node) DependsOn(node Node, opts ...DependencyOption) {
	for _, opt := range opts {
		opt(n, node)
	}

	n.builder.edges = append(n.builder.edges, &adagio.Edge{
		Source:      node.name,
		Destination: n.name,
	})
}
