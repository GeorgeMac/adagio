package debug

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
	"github.com/georgemac/adagio/pkg/worker"
	"github.com/georgemac/adagio/pkg/workflow"
)

const name = "debug"

var (
	_ worker.Function   = (*Function)(nil)
	_ workflow.Function = (*Function)(nil)
)

// Runtime returns the debub packages runtime
func Runtime() worker.Runtime {
	return worker.RuntimeFunc(name, func() worker.Function { return blankFunction() })
}

func blankFunction() *Function {
	c := &Function{Builder: runtime.NewBuilder(name)}

	c.String(&c.Conclusion, "conclusion", false, "success")
	c.Int64(&c.Sleep, "sleep", false, 0)
	c.Strings(&c.Chances, "chances", false)

	return c
}

// NewFunction constructure and configures a new Function pointer
// The function will result in the provided conclusion unless
// one of zero or more chance conditions come true. In the
// case a chance condition is true, the result within the
// condition is made instead.
//
// e.g. With(Chance(0.5, Panic)) will result in the runtime
// causing a panic 50% of the time
func NewFunction(conclusion adagio.Result_Conclusion, opts ...Option) *Function {
	function := blankFunction()
	function.Conclusion = strings.ToLower(conclusion.String())

	Options(opts).Apply(function)

	return function
}

// Function is a structure which contains the arguments
// for a debug package runtime function
type Function struct {
	*runtime.Builder
	Conclusion string
	Sleep      int64
	Chances    []string
}

// Option is a function option for the Function type
type Option func(*Function)

// Options is a slice of Option types
type Options []Option

// Apply functions each option in order on the provided Function
func (o Options) Apply(c *Function) {
	for _, opt := range o {
		opt(c)
	}
}

// WithSleep configures a sleep on the debug function
func WithSleep(dur time.Duration) Option {
	return func(c *Function) {
		c.Sleep = int64(dur)
	}
}

// Result is a string type which reprents the
// result of an operation ran by debug runtime function
type Result string

const (
	// Success is a successful result expectation
	Success Result = "success"
	// Fail is a failure result expectation
	Fail Result = "fail"
	// Error is an error result expectation
	Error Result = "error"
	// Panic is a system panic expecation
	Panic Result = "panic"
)

// ChanceCondition is a structure which contains a
// potential result for the debug function and a probability
// represented as a float64 in the range [0, 1)
type ChanceCondition struct {
	Result      Result
	Probability float64
}

// Chance configures a new ChanceCondition
func Chance(prob float64, res Result) ChanceCondition {
	return ChanceCondition{res, prob}
}

// With configures a set of ChanceConditions on a Function
// when the option is invoked
func With(chances ...ChanceCondition) Option {
	return func(c *Function) {
		for _, chance := range chances {
			c.Chances = append(c.Chances, fmt.Sprintf("%.2f %s", chance.Probability, chance.Result))
		}
	}
}

// Run invokes the desired debug operations.
// It first sleeps the configured amount and
// then loops over any provided chance conditions.
// Given no chance condinition is met it returns the configured
// adagio result conclusion
func (function *Function) Run() (*adagio.Result, error) {
	time.Sleep(time.Duration(function.Sleep))

	for _, c := range function.Chances {
		var chance ChanceCondition
		if _, err := fmt.Sscanf(c, "%f %s", &chance.Probability, &chance.Result); err != nil {
			return nil, err
		}

		if rand.Float64() < chance.Probability {
			switch chance.Result {
			case Panic:
				panic("uh oh")
			case Error:
				return nil, errors.New("debug: error condition")
			default:
				conclusion := strings.ToUpper(string(chance.Result))
				return &adagio.Result{
					Conclusion: adagio.Result_Conclusion(adagio.Result_Conclusion_value[conclusion]),
				}, nil
			}
		}
	}

	conclusion := strings.ToUpper(function.Conclusion)
	return &adagio.Result{
		Conclusion: adagio.Result_Conclusion(adagio.Result_Conclusion_value[conclusion]),
	}, nil
}
