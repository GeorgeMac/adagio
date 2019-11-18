package workflow

import (
	"context"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"google.golang.org/grpc"
)

func function(s *adagio.Node_Spec) Function {
	return FunctionFunc(func(name string) (*adagio.Node_Spec, error) {
		s.Name = name
		return s, nil
	})
}

type FunctionFunc func(string) (*adagio.Node_Spec, error)

func (fn FunctionFunc) NewSpec(name string) (*adagio.Node_Spec, error) {
	return fn(name)
}

type mappable struct {
	FunctionFunc
	argument, input string
}

func Mappable(fn FunctionFunc) *mappable {
	return &mappable{FunctionFunc: fn}
}

func (m *mappable) SetArgumentFromInput(argument, input string) error {
	m.argument, m.input = argument, input
	return nil
}

type client struct {
	controlplane.ControlPlaneClient

	req  *controlplane.StartRequest
	resp *controlplane.StartResponse
}

func (c *client) Start(ctx context.Context, in *controlplane.StartRequest, opts ...grpc.CallOption) (*controlplane.StartResponse, error) {
	c.req = in

	return c.resp, nil
}
