package adagio

import (
	"errors"
	"math/rand"
	"time"

	"github.com/georgemac/adagio/pkg/graph"
	"github.com/oklog/ulid"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func NewRun(spec *GraphSpec) (*Run, error) {
	now := time.Now().UTC()

	run := &Run{
		Id:        ulid.MustNew(ulid.Timestamp(now), entropy).String(),
		CreatedAt: now.Format(time.RFC3339),
		Edges:     spec.Edges,
		Nodes:     buildNodes(spec.Nodes),
	}

	if err := validateGraph(run); err != nil {
		return nil, err
	}

	return run, nil
}

func validateGraph(run *Run) error {
	graph := GraphFrom(run)

	if len(graph.Cycles()) > 0 {
		return errors.New("cannot contain cycles")
	}

	return nil
}

func buildNodes(specs []*Node_Spec) (nodes []*Node) {
	for _, spec := range specs {
		nodes = append(nodes, &Node{
			Spec: spec,
		})
	}

	return
}

func GraphFrom(run *Run) *graph.Graph {
	var (
		graph  = graph.New()
		lookup = map[string]*Node{}
	)

	for _, node := range run.Nodes {
		lookup[node.Spec.Name] = node
		graph.AddNodes(node)
	}

	for _, edge := range run.Edges {
		src := lookup[edge.Source]
		dst := lookup[edge.Destination]
		graph.Connect(src, dst)
	}

	return graph
}
