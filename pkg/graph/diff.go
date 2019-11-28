package graph

import (
	"fmt"
	"sort"

	"github.com/kr/pretty"
)

// Diff returns a string slice of differences between graphs a and b
func Diff(a, b *Graph) []string {
	type graph struct {
		nodes []string
		edges []string
	}

	flatten := func(g *Graph, dest *graph) {
		for src, eset := range g.forward {
			dest.nodes = append(dest.nodes, fmt.Sprintf("%v", src))

			for target := range eset {
				dest.edges = append(dest.edges, fmt.Sprintf("%v to %v", src, target))
			}
		}

		sort.Strings(dest.nodes)
		sort.Strings(dest.edges)
	}

	var ag, bg graph

	flatten(a, &ag)
	flatten(b, &bg)

	return pretty.Diff(ag, bg)
}
