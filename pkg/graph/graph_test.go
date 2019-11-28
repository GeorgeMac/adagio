package graph

import (
	"fmt"
	"sort"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type node struct {
	Name string
}

var (
	a       = node{Name: "a"}
	b       = node{Name: "b"}
	c       = node{Name: "c"}
	d       = node{Name: "d"}
	e       = node{Name: "e"}
	f       = node{Name: "f"}
	g       = node{Name: "g"}
	nodes   = []Node{a, b, c, d, e, f, g}
	example *Graph
)

func init() {
	// example graph connections
	//
	// (a) ---> (c)----
	//   \             \
	//    ------v       v
	//         (d) --> (e) --> (g)
	//    ------^               ^
	//   /                     /
	// (b) --> (f) ------------
	example = New(nodes...)
	example.Connect(a, c)
	example.Connect(a, d)
	example.Connect(b, d)
	example.Connect(b, f)
	example.Connect(c, e)
	example.Connect(d, e)
	example.Connect(e, g)
	example.Connect(f, g)
}

func Test_Graph_Construction(t *testing.T) {
	var (
		actual   = New()
		expected = &Graph{
			forward: edgeSet{
				a: nodeSet{
					b: struct{}{},
					c: struct{}{},
				},
				b: nodeSet{c: struct{}{}},
				c: nodeSet{},
			},
			reverse: edgeSet{
				a: nodeSet{},
				b: nodeSet{a: struct{}{}},
				c: nodeSet{
					a: struct{}{},
					b: struct{}{},
				},
			},
		}
	)

	// (a) -> (b) -> (c)
	//   \------------^
	//
	actual.AddNodes(nodes[0:3]...)
	require.Nil(t, actual.Connect(a, b))
	require.Nil(t, actual.Connect(b, c))
	require.Nil(t, actual.Connect(a, c))

	// ensure you cannot connect to a missing node
	assert.Error(t, actual.Connect(a, d), `node "d": node is not present in graph`)

	if !assert.Equal(t, expected, actual) {
		pretty.Println(actual)
	}

}

func Test_Graph_Edges(t *testing.T) {
	for _, testCase := range []struct {
		Node     node
		Incoming map[Node]struct{}
		Outgoing map[Node]struct{}
	}{
		{
			Node:     a,
			Incoming: map[Node]struct{}{},
			Outgoing: map[Node]struct{}{c: {}, d: {}},
		},
		{
			Node:     b,
			Incoming: map[Node]struct{}{},
			Outgoing: map[Node]struct{}{d: {}, f: {}},
		},
		{
			Node:     c,
			Incoming: map[Node]struct{}{a: {}},
			Outgoing: map[Node]struct{}{e: {}},
		},
		{
			Node:     d,
			Incoming: map[Node]struct{}{a: {}, b: {}},
			Outgoing: map[Node]struct{}{e: {}},
		},
		{
			Node:     e,
			Incoming: map[Node]struct{}{c: {}, d: {}},
			Outgoing: map[Node]struct{}{g: {}},
		},
		{
			Node:     f,
			Incoming: map[Node]struct{}{b: {}},
			Outgoing: map[Node]struct{}{g: {}},
		},
		{
			Node:     g,
			Incoming: map[Node]struct{}{e: {}, f: {}},
			Outgoing: map[Node]struct{}{},
		},
	} {
		t.Run(fmt.Sprintf("%v", testCase.Node), func(t *testing.T) {
			incoming, err := example.Incoming(testCase.Node)
			require.Nil(t, err)
			assert.Equal(t, testCase.Incoming, incoming)
			_, err = example.Incoming(node{"h"})
			assert.Error(t, err, `node "h": node is not present in graph`)

			outgoing, err := example.Outgoing(testCase.Node)
			require.Nil(t, err)
			assert.Equal(t, testCase.Outgoing, outgoing)
			_, err = example.Outgoing(node{"h"})
			assert.Error(t, err, `node "h": node is not present in graph`)
		})
	}
}

func Test_Graph_Cyclic(t *testing.T) {
	// acyclic
	require.Nil(t, example.Cycles(), "graph contains unexpected cycles")

	// (a) -> (b) -> (c)
	//  ^------------/
	//
	cycles := New(a, b, c)
	require.Nil(t, cycles.Connect(a, b))
	require.Nil(t, cycles.Connect(b, c))
	require.Nil(t, cycles.Connect(c, a))

	var (
		expected = [][]Node{{a, b, c}}
		result   = cycles.Cycles()
	)

	require.Len(t, result, 1)

	sort.Slice(result[0], func(i, j int) bool {
		return result[0][i].(node).Name < result[0][j].(node).Name
	})

	assert.Equal(t, expected, result)
}

func Test_Graph_IsRoot(t *testing.T) {
	for _, testCase := range []struct {
		Node   node
		IsRoot bool
	}{
		{Node: a, IsRoot: true},
		{Node: b, IsRoot: true},
		{Node: c},
		{Node: d},
		{Node: e},
		{Node: f},
		{Node: g},
	} {
		caseName := fmt.Sprintf("IsRoot(%q) returns %v", testCase.Node, testCase.IsRoot)
		t.Run(caseName, func(t *testing.T) {
			isRoot, err := example.IsRoot(testCase.Node)
			require.Nil(t, err)
			assert.Equal(t, testCase.IsRoot, isRoot)
		})
	}
}

func Test_Graph_IsLeaf(t *testing.T) {
	for _, testCase := range []struct {
		Node   node
		IsLeaf bool
	}{
		{Node: a},
		{Node: b},
		{Node: c},
		{Node: d},
		{Node: e},
		{Node: f},
		{Node: g, IsLeaf: true},
	} {
		caseName := fmt.Sprintf("IsLeaf(%q) root returns %v", testCase.Node, testCase.IsLeaf)
		t.Run(caseName, func(t *testing.T) {
			isLeaf, err := example.IsLeaf(testCase.Node)
			require.Nil(t, err)
			assert.Equal(t, testCase.IsLeaf, isLeaf)
		})
	}
}
