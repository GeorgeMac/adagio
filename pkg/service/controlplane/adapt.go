package controlplane

import (
	"errors"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/graph"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
)

func toAdagioGraph(input *controlplane.Graph) (adagio.Graph, error) {
	var (
		nodes      []graph.Node
		nameToNode = map[string]*adagio.Node{}
	)

	for _, n := range input.Nodes {
		node := &adagio.Node{Name: n.Name}
		nodes = append(nodes, node)
		nameToNode[node.Name] = node
	}

	graph := graph.New(nodes...)
	for _, e := range input.Edges {
		src, ok := nameToNode[e.Source]
		if !ok {
			return adagio.Graph{}, errors.New("node not defined")
		}

		dest, ok := nameToNode[e.Destination]
		if !ok {
			return adagio.Graph{}, errors.New("node not defined")
		}

		// configure an edge from source to destination
		graph.Connect(src, dest)
	}

	return adagio.NewGraph(graph), nil
}

func toPBRuns(runs ...*adagio.Run) (pbruns []*controlplane.Run) {
	for _, run := range runs {
		pbruns = append(pbruns, toPBRun(run))
	}

	return
}

func toPBRun(run *adagio.Run) *controlplane.Run {
	return &controlplane.Run{Id: run.ID, CreatedAt: run.CreatedAt.Format(time.RFC3339)}
}
