package exec

import (
	"os/exec"

	"github.com/georgemac/adagio/pkg/adagio"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
)

type Runtime struct {
	spec    *runtime.Spec
	command runtime.StringField
	args    runtime.StringsField
}

func New() *Runtime {
	spec := runtime.NewSpec("exec")

	return &Runtime{
		command: spec.String("command", true, ""),
		args:    spec.Strings("args", false, []string{}),
	}
}

func (r *Runtime) Run(n *adagio.Node) (*adagio.Result, error) {
	var (
		command string
		args    []string
	)

	if err := runtime.Parse(n.Spec,
		r.command(&command),
		r.args(&args)); err != nil {

		return nil, err
	}

	data, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &adagio.Result{
		Conclusion: adagio.Result_SUCCESS,
		Output:     data,
	}, nil
}
