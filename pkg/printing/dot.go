package printing

import (
	"fmt"

	"github.com/georgemac/adagio/pkg/adagio"
)

var tableTmpl = `%s [shape=none, margin=0, label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
    <tr><td bgcolor="lightgrey">%s (attempts %d)</td></tr>
    <tr><td bgcolor="%s">%s</td></tr>
</table>>];
`

// FormatterFunc is a function which implements the fmt.Formatter
// interface
type FormatterFunc func(fmt.State, rune)

// Format delegates to the underlying FormatterFunc
func (fn FormatterFunc) Format(f fmt.State, c rune) {
	fn(f, c)
}

// Dot takes an adagio Run proto struct and returns a formatter
func Dot(pbrun *adagio.Run) (fmt.Formatter, error) {
	return FormatterFunc(func(w fmt.State, c rune) {
		fmt.Fprintf(w, "digraph %q {\n", pbrun.Id)
		fmt.Fprintln(w, "rankdir=LR;")

		for _, node := range pbrun.Nodes {
			var (
				color    = "white"
				attempts = len(node.Attempts)
			)

			adagio.VisitLatestAttempt(node, func(result *adagio.Node_Result) {
				switch result.Conclusion {
				case adagio.Node_Result_SUCCESS:
					color = "green"
				case adagio.Node_Result_FAIL:
					color = "red"
				case adagio.Node_Result_ERROR:
					color = "orange"
				default:
					color = "grey"
				}
			})

			fmt.Fprintf(w, tableTmpl, node.Spec.Name, node.Spec.Name, attempts, color, node.Spec.Runtime)
		}

		for _, edge := range pbrun.Edges {
			fmt.Fprintf(w, "    %s -> %s;\n", edge.Source, edge.Destination)
		}

		fmt.Fprint(w, "}")
	}), nil
}
