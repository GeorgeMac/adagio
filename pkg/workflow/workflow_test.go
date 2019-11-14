package workflow

import (
	"context"
	"testing"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/stretchr/testify/assert"
)

func Test_Builder_Simple(t *testing.T) {
	var (
		aSpec = &adagio.Node_Spec{Name: "a"}
		bSpec = &adagio.Node_Spec{Name: "b"}
		cSpec = &adagio.Node_Spec{
			Name: "c",
			Retry: map[string]*adagio.Node_Spec_Retry{
				"fail": &adagio.Node_Spec_Retry{MaxAttempts: 2},
			},
		}
		dSpec = &adagio.Node_Spec{Name: "d"}

		emptySpec = SpecBuilderFunc(func(name string) (*adagio.Node_Spec, error) {
			return &adagio.Node_Spec{Name: name}, nil
		})

		expected = &adagio.Run{Id: "foo"}
		resp     = &controlplane.StartResponse{Run: expected}
		client   = &client{resp: resp}

		expectedGraphSpec = &adagio.GraphSpec{
			Nodes: []*adagio.Node_Spec{
				aSpec,
				bSpec,
				cSpec,
				dSpec,
			},
			Edges: []*adagio.Edge{
				{Source: "a", Destination: "c"},
				{Source: "b", Destination: "c"},
				{Source: "c", Destination: "d"},
			},
		}
	)

	var (
		builder = NewBuilder()

		a = builder.Node("a", emptySpec)
		b = builder.Node("b", emptySpec)
		c = builder.Node("c", emptySpec, WithRetry(adagio.OnFail, 2))
		d = builder.Node("d", emptySpec)
	)

	c.DependsOn(a, b)
	d.DependsOn(c)

	run, err := builder.Start(context.Background(), client)

	assert.Nil(t, err)
	assert.Equal(t, expected, run)
	assert.Equal(t, expectedGraphSpec, client.req.Spec)
}
