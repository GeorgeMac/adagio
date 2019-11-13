package exec

import (
	"os/exec"

	"github.com/georgemac/adagio/pkg/adagio"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
)

// Call is a structure which contains the arguments
// for an exec package runtime call
type Call struct {
	*runtime.Builder
	Command string
	Args    []string
}

// NewCall configures a new exec.Call pointer
func NewCall(command string, args ...string) *Call {
	call := newCall()
	call.Command = command
	call.Args = args
	return call
}

func newCall() *Call {
	c := &Call{Builder: runtime.NewBuilder("exec")}

	c.String("command", true, "")(&c.Command)
	c.Strings("args", false)(&c.Args)

	return c
}

// Runtime is a struct which implements the worker.Runtime
// It executes the work for a provided node on a call to Run
// and uses the os/exec package to invoke a subprocess
type Runtime struct{}

// NewRuntime configures and returns a new Runtime pointer
func NewRuntime() *Runtime {
	return &Runtime{}
}

// Run parses the command and arguments from the provided Node and then
// spawns a subprocess for the desired command and returns the combined
// output writer as an adagio Result output
func (r *Runtime) Run(n *adagio.Node) (*adagio.Result, error) {
	call := newCall()
	call.Parse(n.Spec)

	data, err := exec.Command(call.Command, call.Args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &adagio.Result{
		Conclusion: adagio.Result_SUCCESS,
		Output:     data,
	}, nil
}
