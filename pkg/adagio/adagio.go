package adagio

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/georgemac/adagio/pkg/graph"
)

type (
	Run struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Graph     Graph     `json:"graph"`
	}

	Node struct {
		Name string `json:"-"`
	}

	NodeState string

	Event struct {
		Run      *Run
		Node     *Node
		From, To NodeState
	}

	serializedGraph struct {
		Nodes nodeSet `json:"nodes"`
		Edges edges   `json:"edges"`
	}

	edges   map[string][]string
	nodeSet map[string]*Node
)

var (
	// ErrRunDoesNotExist is returned when an attempt is made to interface
	// with a non existent run
	ErrRunDoesNotExist = errors.New("run does not exist")
	// ErrNodeNotReady is returned when an attempt is made to claim a node in a waiting
	// state
	ErrNodeNotReady = errors.New("node not ready")
)

const (
	NoneState      = NodeState("")
	WaitingState   = NodeState("waiting")
	ReadyState     = NodeState("ready")
	RunningState   = NodeState("running")
	CompletedState = NodeState("completed")
	DeadState      = NodeState("dead")
)

func (n Node) String() string {
	return fmt.Sprintf("(%s)", n.Name)
}

type Graph struct {
	graph *graph.Graph
}

func NewGraph(graph *graph.Graph) Graph {
	return Graph{graph}
}

func (g *Graph) IsRoot(n *Node) (bool, error) {
	return g.graph.IsRoot(n)
}

func (g *Graph) Incoming(n *Node) (map[*Node]struct{}, error) {
	incoming, err := g.graph.Incoming(n)
	if err != nil {
		return nil, err
	}

	return convertMap(incoming), nil
}

func (g *Graph) Outgoing(n *Node) (map[*Node]struct{}, error) {
	outgoing, err := g.graph.Outgoing(n)
	if err != nil {
		return nil, err
	}

	return convertMap(outgoing), nil
}

func convertMap(nodes map[graph.Node]struct{}) map[*Node]struct{} {
	dest := map[*Node]struct{}{}

	for node, _ := range nodes {
		dest[node.(*Node)] = struct{}{}
	}

	return dest
}

type VisitFunc func(*Node)

func (g *Graph) Walk(fn VisitFunc) {
	g.graph.Walk(graph.Backwards, func(n graph.Node) {
		fn(n.(*Node))
	})
}

func (g *Graph) UnmarshalJSON(v []byte) error {
	var sGraph serializedGraph
	if err := json.Unmarshal(v, &sGraph); err != nil {
		return err
	}

	g.graph = graph.New()

	for name, node := range sGraph.Nodes {
		// copy name from key to node object
		node.Name = name
		// add node to graph
		g.graph.AddNodes(node)
	}

	for name, edges := range sGraph.Edges {
		// TODO validate presence of node in map
		from := sGraph.Nodes[name]

		for _, destName := range edges {
			// TODO validate presence of node in map
			to := sGraph.Nodes[destName]

			// create connection from source node to destination node
			g.graph.Connect(from, to)
		}
	}

	return nil
}
