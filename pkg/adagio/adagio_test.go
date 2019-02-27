package adagio

import (
	"encoding/json"
	"testing"

	"github.com/georgemac/adagio/pkg/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraph_UnmarshalJSON(t *testing.T) {
	var (
		nodes = []graph.Node{
			&Node{Name: "a"},
			&Node{Name: "b"},
			&Node{Name: "c"},
			&Node{Name: "d"},
			&Node{Name: "e"},
			&Node{Name: "f"},
			&Node{Name: "g"},
		}
		expGraph = graph.New(nodes...)
		actual   Graph
	)

	expGraph.Connect(nodes[0], nodes[2])
	expGraph.Connect(nodes[0], nodes[3])
	expGraph.Connect(nodes[1], nodes[3])
	expGraph.Connect(nodes[1], nodes[5])
	expGraph.Connect(nodes[2], nodes[4])
	expGraph.Connect(nodes[3], nodes[4])
	expGraph.Connect(nodes[4], nodes[6])
	expGraph.Connect(nodes[5], nodes[6])

	require.Nil(t, json.Unmarshal(simpleSchema, &actual))

	diff := graph.Diff(expGraph, actual.graph)
	assert.Len(t, diff, 0, diff)
}

var (
	simpleSchema = []byte(`{
    "nodes": {
        "a": {},
        "b": {},
        "c": {},
        "d": {},
        "e": {},
        "f": {},
        "g": {}
    },
    "edges": {
        "a": ["c", "d"],
        "b": ["d", "f"],
        "c": ["e"],
        "d": ["e"],
        "e": ["g"],
        "f": ["g"]
    }
}`)
)
