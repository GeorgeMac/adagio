// The runtime package contains types which aid in the building of runtimes
// and runtime calls which can be constructed and invoked by the worker types
//
// The *Builder type aids in composing types which avoid having to do fiddly
// parsing to and from metadata fields on adagio.Node types
//
// The following is a contrived example of implementing the Runtime and Call types
// used the *Builder helper type
//
// package thing
//
// import "github.com/georgemac/adagio/pkg/worker"
//
// const name = "thing"
//
// type Runtime struct {}
//
// func (r Runtime) Name() string { return name }
//
// func (r Runtime) BlankCall() worker.Call {
//     return blankCall()
// }
//
// func blankCall() *Call {
//     call := &Call{Builder: runtime.New(name)}
//
//     call.String("string_arg", true, "default")(&call.StringArg)
//     call.Strings("strings_arg", false, "many", "default")(&call.StringsArg)
//
//     return call
// }
//
// type Call struct {
//     *runtime.Builder
//     StringArg  string
//     StringsArg []string
// }
//
// func NewCall(stringArg string, extraStringArgs ...string) *Call {
//     call := blackCall()
//     call.StringArg = stringArg
//     call.StringsArg = extraStringArgs
//
//     return call
// }
//
// func (c *Call) Run() error {
//     // do the things with the arguments
// }
package runtime
