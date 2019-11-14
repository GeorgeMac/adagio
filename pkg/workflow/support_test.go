package workflow

import (
	"context"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"google.golang.org/grpc"
)

func spec(s *adagio.Node_Spec) SpecBuilder {
	return SpecBuilderFunc(func(name string) (*adagio.Node_Spec, error) {
		s.Name = name
		return s, nil
	})
}

type SpecBuilderFunc func(string) (*adagio.Node_Spec, error)

func (s SpecBuilderFunc) NewSpec(name string) (*adagio.Node_Spec, error) {
	return s(name)
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
