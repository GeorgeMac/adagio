// Package runtime contains types which aid in the building of runtimes
// and runtime calls which can be constructed and invoked by the agent types.
//
// The *Builder type aids in composing types which avoid having to do manual
// parsing to and from metadata fields on adagio.Node types.
//
// The `runtime.Function()` is a handy function which takes a type with the behavior for
// parsing an incoming node (`Parse(*adagio.Node) error`) and running the function once
// parsed (`Run() (*adagio.Result, error)`) as separate calls and combines them in a agent.Function
// compatible implementation.
// This allows for a *runtime.Builder to be embedded into new Function definitions,
// which can then take on the responsibility of the `Parse(node) error` call.
// All you need to do is define and configure the named arguments of your function using
// the helper methods on the builder and then a single `Run()` implementation which takes
// care of invoking your functions behavior.
// For example, `builder.String(&str, "foo", false, "bar")` will add an argument "foo" which
// is not required and defaults to the value "bar".
//
// The following is a contrived example of implementing the Runtime and Function types
// used the *Builder helper type.
//
// package thing
//
// import "github.com/georgemac/adagio/pkg/agent"
//
// const name = "thing"
//
// type Runtime struct {}
//
// func (r Runtime) Name() string { return name }
//
// func (r Runtime) BlankFunction() agent.Function {
//     return runtime.Function(blankFunction())
// }
//
// func blankFunction() *Function {
//     call := &Function{Builder: runtime.New(name)}
//
//     call.String("string_arg", true, "default")(&call.StringArg)
//     call.Strings("strings_arg", false, "many", "default")(&call.StringsArg)
//
//     return call
// }
//
// type Function struct {
//     *runtime.Builder
//     StringArg  string
//     StringsArg []string
// }
//
// func NewFunction(stringArg string, extraStringArgs ...string) *Function {
//     call := blackFunction()
//     call.StringArg = stringArg
//     call.StringsArg = extraStringArgs
//
//     return call
// }
//
// func (c *Function) Run() error {
//     // do the things with the arguments
// }
package runtime
