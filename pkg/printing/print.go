package printing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/kr/pretty"
)

type (
	Node struct {
		Name       string
		Runtime    string
		Metadata   map[string][]string
		Status     string
		Conclusion string
		StartedAt  time.Time
		FinishedAt time.Time
	}

	Run struct {
		ID        string
		CreatedAt time.Time
		Nodes     []Node
	}
)

// PBRunToRun takes a protobuf run and adapts it to the printer.Run type
// which is used to produce a friendlier printing output
func PBRunToRun(pbrun *adagio.Run) (Run, error) {
	var (
		createdAt, _ = time.Parse(time.RFC3339, pbrun.CreatedAt)
		run          = Run{
			ID:        pbrun.Id,
			CreatedAt: createdAt,
		}
	)

	for _, node := range pbrun.Nodes {
		var (
			conclusion    adagio.Node_Result_Conclusion
			startedAt, _  = time.Parse(time.RFC3339, node.StartedAt)
			finishedAt, _ = time.Parse(time.RFC3339, node.FinishedAt)
			status, err   = statusToString(node.Status)
		)
		if err != nil {
			return Run{}, err
		}

		adagio.VisitLatestAttempt(node, func(result *adagio.Node_Result) {
			conclusion = result.Conclusion
		})

		var metadata map[string][]string
		if len(node.Spec.Metadata) > 0 {
			metadata = map[string][]string{}

			for k, v := range node.Spec.Metadata {
				var values []string

				for _, value := range v.Values {
					values = append(values, value)
				}

				metadata[k] = values
			}
		}

		run.Nodes = append(run.Nodes, Node{
			Name:       node.Spec.Name,
			Runtime:    node.Spec.Runtime,
			Metadata:   metadata,
			Status:     status,
			Conclusion: conclusionToString(conclusion),
			StartedAt:  startedAt,
			FinishedAt: finishedAt,
		})
	}

	return run, nil
}

func Spew(pbrun *adagio.Run) (fmt.Formatter, error) {
	run, err := PBRunToRun(pbrun)
	if err != nil {
		return nil, err
	}

	return spew.NewFormatter(run), nil
}

func Pretty(pbrun *adagio.Run) (fmt.Formatter, error) {
	run, err := PBRunToRun(pbrun)
	if err != nil {
		return nil, err
	}

	return pretty.Formatter(run), nil
}

func statusToString(state adagio.Node_Status) (string, error) {
	switch state {
	case adagio.Node_WAITING:
		return "waiting", nil
	case adagio.Node_READY:
		return "ready", nil
	case adagio.Node_RUNNING:
		return "running", nil
	case adagio.Node_COMPLETED:
		return "completed", nil
	default:
		return "", errors.New("state not recognized")
	}
}

func conclusionToString(conclusion adagio.Node_Result_Conclusion) string {
	return strings.ToLower(conclusion.String())
}
