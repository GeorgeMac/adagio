package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/workflow"
)

var _ workflow.Function = (*Builder)(nil)

// ArgumentType is a type of argument
type ArgumentType uint8

// String prints a string representation of the argument type
func (t ArgumentType) String() string {
	switch t {
	case StringArgumentType:
		return "string"
	case StringsArgumentType:
		return "[]string"
	case Int64ArgumentType:
		return "int64"
	case TimeArgumentType:
		return "time.Time"
	default:
		return "unknown"
	}
}

const (
	// StringArgumentType represents a string type argument
	StringArgumentType ArgumentType = iota
	// StringsArgumentType represents a slice of string types argument
	StringsArgumentType
	// Int64ArgumentType represents an int64 type argument
	Int64ArgumentType
	// TimeArgumentType represent a time.Time type argument
	TimeArgumentType
)

// ParseRunner is a type which has a separate function for parsing a node
// and then invoking the desired behavior via Run
type ParseRunner interface {
	Parse(n *adagio.Node) error
	Run() (*adagio.Result, error)
}

// FunctionAdaptor is the type used to covert a ParseRunner
// into a worker.Function compatable type
type FunctionAdaptor struct {
	runner ParseRunner
}

// Run calls Parse on the provided node and then invokes
// Run on the underlying ParseRunner
func (p FunctionAdaptor) Run(n *adagio.Node) (*adagio.Result, error) {
	if err := p.runner.Parse(n); err != nil {
		return nil, err
	}

	return p.runner.Run()
}

// Function converts the provided ParseRunner into a worker.Function
// using the FunctionAdaptor wrapper type
func Function(runner ParseRunner) FunctionAdaptor {
	return FunctionAdaptor{runner}
}

// Builder is a struct aids in both parsing and construction
// of runtimes and runtime request types
type Builder struct {
	name      string
	arguments map[string]*Argument
}

// NewBuilder constructs and configures a new runtime builder
func NewBuilder(name string) *Builder {
	return &Builder{name: name, arguments: map[string]*Argument{}}
}

// Name returns the name of the runtime being built
func (p *Builder) Name() string {
	return p.name
}

// SetArgumentFromInput configures the argument to derives its value from
// the input paramertes of the node
func (b *Builder) SetArgumentFromInput(argument, name string) error {
	f, ok := b.arguments[argument]
	if !ok {
		return fmt.Errorf("argument not found %q", argument)
	}

	f.fromInput = name

	return nil
}

// NewSpec constructs a Node_Spec based on the current state
// of the builders arguments
func (p *Builder) NewSpec(name string) (*adagio.Node_Spec, error) {
	spec := &adagio.Node_Spec{
		Name:     name,
		Runtime:  p.name,
		Metadata: map[string]*adagio.MetadataValue{},
	}

	for _, argument := range p.arguments {
		if err := argument.addTo(p.name, spec); err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// Parse parses the state from a node into the targets
// set on the builders arguments
func (p *Builder) Parse(node *adagio.Node) error {
	for _, argument := range p.arguments {
		if err := argument.parseNode(p.name, node); err != nil {
			return err
		}
	}

	return nil
}

// String configures a string argument which when call with a string pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.NewSpec()
func (s *Builder) String(v *string, name string, required bool, defaultValue string) {
	argument := newArgument(name, StringArgumentType, required, []string{defaultValue})
	argument.asMetadata = func() ([]string, error) {
		if v == nil {
			if required {
				return nil, errors.New("argument is required")
			}

			return nil, nil
		}

		return []string{*v}, nil
	}

	argument.parse = func(vs []string) error {
		if len(vs) < 1 {
			return errors.New("no value set for key")
		}

		*v = vs[0]

		return nil
	}

	s.arguments[name] = argument
}

// Strings configures a string slice argument which when call with a string slice pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.NewSpec()
func (s *Builder) Strings(v *[]string, name string, required bool, defaultValues ...string) {
	argument := newArgument(name, StringsArgumentType, required, defaultValues)
	argument.asMetadata = func() ([]string, error) {
		if v == nil {
			if required {
				return nil, errors.New("argument is required")
			}

			return nil, nil
		}

		return *v, nil
	}

	argument.parse = func(vs []string) error {
		*v = vs

		return nil
	}

	s.arguments[name] = argument
}

// Int64 configures an int64 argument which will set the pointer on calls
// to builder.Parse() and read the value at the end of the pointer
// on calls to builder.NewSpec()
func (s *Builder) Int64(v *int64, name string, required bool, defaultValue int64) {
	argument := newArgument(name, Int64ArgumentType, required, []string{fmt.Sprintf("%d", defaultValue)})
	argument.asMetadata = func() ([]string, error) {
		if v == nil {
			if required {
				return nil, errors.New("argument is required")
			}

			return nil, nil
		}

		return []string{fmt.Sprintf("%d", *v)}, nil
	}

	argument.parse = func(vs []string) (err error) {
		if len(vs) < 1 {
			return errors.New("no value set for key")
		}

		*v, err = strconv.ParseInt(vs[0], 10, 64)

		return err
	}

	s.arguments[name] = argument
}

// Time configures a time argument which when call with a int64 pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.NewSpec()
func (s *Builder) Time(v *time.Time, name string, required bool, defaultValue time.Time) {
	argument := newArgument(name, TimeArgumentType, required, []string{defaultValue.Format(time.RFC3339Nano)})
	argument.asMetadata = func() ([]string, error) {
		if v == nil {
			if required {
				return nil, errors.New("argument is required")
			}

			return nil, nil
		}

		return []string{v.Format(time.RFC3339Nano)}, nil
	}

	argument.parse = func(vs []string) (err error) {
		if len(vs) < 1 {
			return errors.New("no value set for key")
		}

		*v, err = time.Parse(time.RFC3339Nano, vs[0])

		return err
	}

	s.arguments[name] = argument
}

// Argument is a structure which contains the properties
// of a argument used to call a runtime
type Argument struct {
	Name     string
	Type     ArgumentType
	Required bool
	Defaults []string

	target     interface{}
	fromInput  string
	parse      func([]string) error
	asMetadata func() ([]string, error)
}

func metadataArgument(runtime string, argument *Argument) string {
	return fmt.Sprintf("adagio.arguments.%s.%s", runtime, argument.Name)
}

func inputArgument(runtime string, argument *Argument) string {
	return fmt.Sprintf("adagio.inputs.%s.%s", runtime, argument.Name)
}

func (f *Argument) addTo(runtime string, spec *adagio.Node_Spec) error {
	if f.fromInput != "" {
		spec.Metadata[inputArgument(runtime, f)] = &adagio.MetadataValue{Values: []string{f.fromInput}}

		return nil
	}

	if f.asMetadata == nil {
		return fmt.Errorf("target not set for %q", f.Name)
	}

	values, err := f.asMetadata()
	if err != nil {
		return err
	}

	spec.Metadata[metadataArgument(runtime, f)] = &adagio.MetadataValue{Values: values}

	return nil
}

// Parse sets the values of the argument using the provided slice of strings
func (f *Argument) parseNode(runtime string, node *adagio.Node) error {
	if f.parse == nil {
		return fmt.Errorf("target not fet for %q", f.Name)
	}

	if err := f.SetToDefault(); err != nil {
		return err
	}

	inputName, ok := node.Spec.Metadata[inputArgument(runtime, f)]
	if ok && len(inputName.Values) > 0 {
		value, ok := node.Inputs[inputName.Values[0]]
		if !ok {
			return f.missing()
		}

		return f.parse([]string{string(value)})
	}

	values, ok := node.Spec.Metadata[metadataArgument(runtime, f)]
	if !ok {
		return f.missing()
	}

	return f.parse(values.Values)
}

func (f *Argument) missing() error {
	if f.Required {
		return fmt.Errorf("missing required argument %q", f.Name)
	}

	return nil
}

// SetToDefault applies the default values to the argument
func (f *Argument) SetToDefault() error {
	return f.parse(f.Defaults)
}

func newArgument(name string, argumentType ArgumentType, required bool, defaultValues []string) *Argument {
	return &Argument{
		Name:     name,
		Type:     argumentType,
		Required: required,
		Defaults: defaultValues,
	}
}
