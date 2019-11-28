package adagio

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/georgemac/adagio/pkg/graph"
	"github.com/oklog/ulid/v2"
)

var (
	entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	mu      sync.Mutex
)

// NewRun converts a graph specification into a new run instance
// This is a convention and helper function for repository implementations to use to
// correctly adapt a new graph spec into a run. It validates that the graph has
// no cycles and initializes states, timestamps and IDs appropriately
func NewRun(spec *GraphSpec) (run *Run, err error) {
	func() {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now().UTC()
		run = &Run{
			Id:        ulid.MustNew(ulid.Timestamp(now), entropy).String(),
			CreatedAt: now.Format(time.RFC3339Nano),
			Edges:     spec.Edges,
			Nodes:     buildNodes(spec.Nodes),
		}
	}()

	graph := GraphFrom(run)

	if err = validateGraph(graph); err != nil {
		return
	}

	err = setInitialNodeStates(graph, run.Nodes)

	return
}

// GetNodeByName fetches a Node from the Run by name
func (run *Run) GetNodeByName(name string) (*Node, error) {
	for _, node := range run.Nodes {
		if node.Spec.Name == name {
			return node, nil
		}
	}

	return nil, errors.New("graph: node not found")
}

func buildNodes(specs []*Node_Spec) (nodes []*Node) {
	for _, spec := range specs {
		nodes = append(nodes, &Node{
			Spec: spec,
		})
	}

	return
}

func validateGraph(graph *graph.Graph) error {
	if len(graph.Cycles()) > 0 {
		return errors.New("cannot contain cycles")
	}

	return nil
}

func setInitialNodeStates(graph *graph.Graph, nodes []*Node) error {
	for _, node := range nodes {
		incoming, err := graph.Incoming(node)
		if err != nil {
			return err
		}

		if len(incoming) == 0 {
			node.Status = Node_READY
			continue
		}

		node.Status = Node_WAITING
	}

	return nil
}

// GraphFrom takes a run and builds a *graph.Graph from it which contains
// helpful functions to traversing the runs graph structure
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
