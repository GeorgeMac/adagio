package debug

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	runtime "github.com/georgemac/adagio/pkg/runtimes"
)

// Call is a structure which contains the arguments
// for a debug package runtime call
type Call struct {
	*runtime.Builder
	Conclusion string
	Sleep      int64
	Chances    []string
}

// Option is a function option for the Call type
type Option func(*Call)

// Options is a slice of Option types
type Options []Option

// Apply calls each option in order on the provided Call
func (o Options) Apply(c *Call) {
	for _, opt := range o {
		opt(c)
	}
}

// WithSleep configures a sleep on the debug call
func WithSleep(dur time.Duration) Option {
	return func(c *Call) {
		c.Sleep = int64(dur)
	}
}

// Result is a string type which reprents the
// result of an operation ran by debug runtime call
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
// potential result for the debug call and a probability
// represented as a float64 in the range [0, 1)
type ChanceCondition struct {
	Result      Result
	Probability float64
}

// Chance configures a new ChanceCondition
func Chance(prob float64, res Result) ChanceCondition {
	return ChanceCondition{res, prob}
}

// With configures a set of ChanceConditions on a Call
// when the option is invoked
func With(chances ...ChanceCondition) Option {
	return func(c *Call) {
		for _, chance := range chances {
			c.Chances = append(c.Chances, fmt.Sprintf("%.2f %s", chance.Probability, chance.Result))
		}
	}
}

// NewCall constructure and configures a new Call pointer
// The call will result in the provided conclusion unless
// one of zero or more chance conditions come true. In the
// case a chance condition is true, the result within the
// condition is made instead.
//
// e.g. With(Chance(0.5, Panic)) will result in the runtime
// causing a panic 50% of the time
func NewCall(conclusion adagio.Result_Conclusion, opts ...Option) *Call {
	call := newCall()
	call.Conclusion = strings.ToLower(conclusion.String())

	Options(opts).Apply(call)

	return call
}

func newCall() *Call {
	c := &Call{Builder: runtime.NewBuilder("debug")}

	c.String("conclusion", false, "success")(&c.Conclusion)
	c.Int64("sleep", false, 0)(&c.Sleep)
	c.Strings("chances", false)(&c.Chances)

	return c
}

// Runtime is a struct which implements the worker.Runtime
// It executes a specification for a debug task. This task
// mostly reports back whatever it is told, along with sleeping
// for a configured amount of time.
type Runtime struct{}

// NewRuntime configures and returns a new Runtime pointer
func NewRuntime() *Runtime {
	return &Runtime{}
}

// Run parse the debug run from the provided Node and then
// invokes the desired operations. It first sleeps the configured ammount
// and then loops over any provided chance conditions.
// Given no chance condinition is met it returns the configured
// adagio result conclusion
func (r *Runtime) Run(n *adagio.Node) (*adagio.Result, error) {
	call := newCall()
	call.Parse(n.Spec)

	time.Sleep(time.Duration(call.Sleep))

	for _, c := range call.Chances {
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

	conclusion := strings.ToUpper(call.Conclusion)
	return &adagio.Result{
		Conclusion: adagio.Result_Conclusion(adagio.Result_Conclusion_value[conclusion]),
	}, nil
}
