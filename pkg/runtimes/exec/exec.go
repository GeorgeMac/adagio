package exec

import (
	"os/exec"

	"github.com/georgemac/adagio/pkg/adagio"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/georgemac/adagio/pkg/workflow"
)

const name = "exec"

var (
	_ worker.Call          = (*Call)(nil)
	_ workflow.SpecBuilder = (*Call)(nil)
)

// Runtime returns the exec package worker.Runtime
func Runtime() worker.Runtime {
	return worker.RuntimeFunc(name, func() worker.Call { return blankCall() })
}

func blankCall() *Call {
	c := &Call{Builder: runtime.NewBuilder(name)}

	c.String("command", true, "")(&c.Command)
	c.Strings("args", false)(&c.Args)

	return c
}

// Call is a struct which implements the worker.Runtime
// It executes the work for a provided node on a call to Run
// and uses the os/exec package to invoke a subprocess
type Call struct {
	*runtime.Builder
	Command string
	Args    []string
}

// NewCall configures a new exec.Call pointer
func NewCall(command string, args ...string) *Call {
	call := blankCall()
	call.Command = command
	call.Args = args
	return call
}

// Run parses the command and arguments from the provided Node and then
// spawns a subprocess for the desired command and returns the combined
// output writer as an adagio Result output
func (call *Call) Run() (*adagio.Result, error) {
	data, err := exec.Command(call.Command, call.Args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &adagio.Result{
		Conclusion: adagio.Result_SUCCESS,
		Output:     data,
	}, nil
}
