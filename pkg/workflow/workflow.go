package workflow

import (
	"context"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

// SpecBuilder is any type which can produce an adagio Node_Spec
// pointer
type SpecBuilder interface {
	NewSpec(name string) (*adagio.Node_Spec, error)
}

// Option is a functional option for a node spec
type Option func(*adagio.Node_Spec)

// Options is a slice of Option types
type Options []Option

// Apply calls each option in turn on the provided Node_Spec
func (o Options) Apply(spec *adagio.Node_Spec) {
	for _, opt := range o {
		opt(spec)
	}
}

// WithRetry configures a retry for a specified condition up to
// maxAttempts times on a Node_Spec
func WithRetry(condition adagio.RetryCondition, maxAttempts int32) Option {
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
	spec *adagio.GraphSpec
	err  error
}

// NewBuilder creates and configures a new Builder
func NewBuilder() Builder {
	return Builder{spec: &adagio.GraphSpec{}}
}

// MustNode panics if err is nil otherwise it returns the provided node
func MustNode(n Node, err error) Node {
	if err != nil {
		panic(err)
	}
	return n
}

// Node creates a node from a provided name and SpecBuilder type
// it also applies any provided options
func (b Builder) Node(name string, s SpecBuilder, opts ...Option) (n Node) {
	if b.err != nil {
		return
	}

	n.spec, b.err = s.NewSpec(name)
	if b.err != nil {
		return
	}

	n.builder = b

	Options(opts).Apply(n.spec)

	b.spec.Nodes = append(b.spec.Nodes, n.spec)

	return
}

// Start invokes the built graph specification onto the provided client
// It returns the run responded by the controlplane API
func (b Builder) Start(ctx context.Context, client controlplane.ControlPlaneClient) (*adagio.Run, error) {
	resp, err := client.Start(ctx, &controlplane.StartRequest{Spec: b.spec})
	if err != nil {
		return nil, err
	}

	return resp.Run, nil
}

// Node is a builder wrapper type which can be used to further
// create connections between nodes in the originating builder
type Node struct {
	builder Builder
	spec    *adagio.Node_Spec
}

// DependsOn creates a connection from the provided nodes (sources)
// to the callee node (destination) on the original builder
func (n Node) DependsOn(nodes ...Node) {
	// escape early if error on builder
	if n.builder.err != nil {
		return
	}

	for _, v := range nodes {
		n.builder.spec.Edges = append(n.builder.spec.Edges, &adagio.Edge{
			Source:      v.spec.Name,
			Destination: n.spec.Name,
		})
	}
}
