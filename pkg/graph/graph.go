package graph

import (
	"github.com/pkg/errors"
)

var (
	// ErrDuplicateNode is returned when a node is attempted to
	// be inserted when it is already present
	ErrDuplicateNode = errors.New("node already exists in graph")
	// ErrMissingNode is returned when a node is reference which
	// does not exist within the graph
	ErrMissingNode = errors.New("node is not present in graph")
)

type (
	// Node is a vertex in a graph
	Node    interface{}
	nodeSet map[Node]struct{}
	edgeSet map[Node]nodeSet
)

// Graph is a graph datastructure which contains nodes with directed
// edges between them
type Graph struct {
	forward, reverse edgeSet
}

// New constructs a new graph from a set of Nodes
func New(nodes ...Node) *Graph {
	graph := &Graph{
		forward: edgeSet{},
		reverse: edgeSet{},
	}

	graph.AddNodes(nodes...)

	return graph
}

func (g *Graph) clone() *Graph {
	return &Graph{
		forward: g.forward.clone(),
		reverse: g.reverse.clone(),
	}
}

func (g *Graph) initializeIfEmpty() {
	if g.forward == nil {
		g.forward = edgeSet{}
	}

	if g.reverse == nil {
		g.reverse = edgeSet{}
	}
}

// AddNodes adds the provided nodes to the graph
func (g *Graph) AddNodes(nodes ...Node) error {
	g.initializeIfEmpty()

	for _, n := range nodes {
		if err := g.forward.addNode(n); err != nil {
			return err
		}

		if err := g.reverse.addNode(n); err != nil {
			return err
		}
	}

	return nil
}

// IsRoot returns true if the node has no incoming edges
// within the graph
func (g *Graph) IsRoot(n Node) (bool, error) {
	if !g.reverse.present(n) {
		return false, wrapMissingErr(n)
	}

	return len(g.reverse[n]) == 0, nil
}

// IsLeaf returns true if the node has no outgoing edges
// within the graph
func (g *Graph) IsLeaf(n Node) (bool, error) {
	if !g.forward.present(n) {
		return false, wrapMissingErr(n)
	}

	return len(g.forward[n]) == 0, nil
}

// Outgoing returns a set of nodes which are directed to from
// the provided node
func (g *Graph) Outgoing(n Node) (nodes map[Node]struct{}, err error) {
	if !g.forward.present(n) {
		return nil, wrapMissingErr(n)
	}

	return g.forward[n].clone(), nil
}

// Incoming returns a set of nodes which are directed from the
// provided node
func (g *Graph) Incoming(n Node) (nodes map[Node]struct{}, err error) {
	if !g.reverse.present(n) {
		return nil, wrapMissingErr(n)
	}

	return g.reverse[n].clone(), nil
}

// Connect constructs and edge from one node to another
func (g *Graph) Connect(from, to Node) error {
	g.initializeIfEmpty()

	if err := g.forward.connect(from, to); err != nil {
		return err
	}

	return g.reverse.connect(to, from)
}

// Cycles detects and returns any cycles found within
// the graph
func (g *Graph) Cycles() (cycles [][]Node) {
	for _, cc := range g.StronglyConnectedComponents() {
		if len(cc) > 1 {
			cycles = append(cycles, cc)
		}
	}

	return
}

// StronglyConnectedComponents returns the set of strongly connected
// components within the graph
func (g *Graph) StronglyConnectedComponents() (components [][]Node) {
	var (
		sorted = g.TopologicalSort()
		set    = g.reverse.clone()
	)

	for _, node := range sorted {
		var dest []Node
		dest = dfs(set, node, dest)
		components = append(components, dest)
	}

	return
}

// TopologicalSort returns the set of nodes within the graph topologically
// sorted
func (g *Graph) TopologicalSort() []Node {
	return reverse(g.dfsForward())
}

// VisitFunc is a function which takes a Node and returns an error
type VisitFunc func(Node) error

// Walk calls each visit func on each node in depth first order
func (g *Graph) Walk(fn VisitFunc) {
	for _, node := range g.dfsForward() {
		fn(node)
	}
}

// WalkFrom traverses the outgoing nodes to the leaves of the graph
// from the provided node
func (g *Graph) WalkFrom(node Node, fn VisitFunc) error {
	var (
		set           = g.forward.clone()
		toExplore, ok = set.remove(node)
	)

	if !ok {
		return wrapMissingErr(node)
	}

	for node := range toExplore {
		if err := walkFrom(set, node, fn); err != nil {
			return err
		}
	}

	return nil
}

func walkFrom(set edgeSet, node Node, fn VisitFunc) error {
	toExplore, ok := set.remove(node)
	if !ok {
		return nil
	}

	if err := fn(node); err != nil {
		return err
	}

	for node := range toExplore {
		if err := walkFrom(set, node, fn); err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) dfsForward() []Node {
	var (
		nodes = make([]Node, 0, len(g.forward))
		set   = g.forward.clone()
	)

	for node := range set {
		nodes = dfs(set, node, nodes)
	}

	return nodes
}

func reverse(nodes []Node) []Node {
	for i := 0; i < len(nodes)/2; i++ {
		opp := len(nodes) - 1 - i
		nodes[i], nodes[opp] = nodes[opp], nodes[i]
	}

	return nodes
}

func dfs(all edgeSet, from Node, dest []Node) []Node {
	targets, ok := all.remove(from)
	if !ok {
		return dest
	}

	for node := range targets {
		dest = dfs(all, node, dest)
	}

	return append(dest, from)
}

func (e edgeSet) present(n Node) (present bool) {
	_, present = e[n]
	return
}

func (e edgeSet) connect(from, to Node) error {
	if !e.present(from) {
		return wrapMissingErr(from)
	}

	if !e.present(to) {
		return wrapMissingErr(to)
	}

	e[from][to] = struct{}{}

	return nil
}

func (e edgeSet) addNode(n Node) error {
	if e.present(n) {
		return errors.Wrapf(ErrDuplicateNode, "node %q", n)
	}

	e[n] = nodeSet{}

	return nil
}

func (e edgeSet) remove(n Node) (set nodeSet, ok bool) {
	set, ok = e[n]
	if ok {
		delete(e, n)
	}

	return
}

func (e edgeSet) clone() edgeSet {
	dest := edgeSet{}
	for n, edge := range e {
		dest[n] = edge
	}

	return dest
}

func (s nodeSet) clone() nodeSet {
	dest := nodeSet{}
	for k, v := range s {
		dest[k] = v
	}

	return dest
}

func wrapMissingErr(n Node) error {
	return errors.Wrapf(ErrMissingNode, "node %q", n)
}
