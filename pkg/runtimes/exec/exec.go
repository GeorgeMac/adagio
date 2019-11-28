package exec

import (
	"os/exec"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/agent"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
	"github.com/georgemac/adagio/pkg/workflow"
)

const name = "exec"

var (
	_ workflow.Function = (*Function)(nil)
)

// Runtime returns the exec package agent.Runtime
func Runtime() agent.Runtime {
	return agent.RuntimeFunc(name, func() agent.Function {
		return runtime.Function(blankFunction())
	})
}

func blankFunction() *Function {
	c := &Function{Builder: runtime.NewBuilder(name)}

	c.String(&c.Command, "command", true, "")
	c.Strings(&c.Args, "args", false)

	return c
}

// Function is a struct which implements the agent.Runtime
// It executes the work for a provided node on a function to Run
// and uses the os/exec package to invoke a subprocess
type Function struct {
	*runtime.Builder
	Command string
	Args    []string
}

// NewFunction configures a new exec.Function pointer
func NewFunction(command string, args ...string) *Function {
	fn := blankFunction()
	fn.Command = command
	fn.Args = args
	return fn
}

// Run spawns a subprocess for the desired command and returns the combined
// output writer as an adagio Result output slice of bytes
func (fn *Function) Run() (*adagio.Result, error) {
	data, err := exec.Command(fn.Command, fn.Args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &adagio.Result{
		Conclusion: adagio.Result_SUCCESS,
		Output:     data,
	}, nil
}
