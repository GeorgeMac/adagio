package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
)

// FieldType is a type of field
type FieldType uint8

// String prints a string representation of the field type
func (t FieldType) String() string {
	switch t {
	case StringFieldType:
		return "string"
	case StringsFieldType:
		return "[]string"
	case Int64FieldType:
		return "int64"
	default:
		return "unknown"
	}
}

const (
	// StringFieldType represents a string type field
	StringFieldType FieldType = iota
	// StringsFieldType represents a slice of string types field
	StringsFieldType
	// Int64FieldType represents an int64 type field
	Int64FieldType
	// TimeFieldType represent a time.Time type field
	TimeFieldType
)

// Builder is a struct aids in both parsing and construction
// of runtimes and runtime request types
type Builder struct {
	name   string
	fields []*Field
}

// NewBuilder constructs and configures a new runtime builder
func NewBuilder(name string) *Builder {
	return &Builder{name: name}
}

// Name returns the name of the runtime being built
func (p *Builder) Name() string {
	return p.name
}

func (p *Builder) fieldName(field *Field) string {
	return fmt.Sprintf("adagio.runtime.%s.%s", p.name, field.Name)
}

// Spec constructs a Node_Spec based on the current state
// of the builders fields
func (p *Builder) Spec() (*adagio.Node_Spec, error) {
	spec := &adagio.Node_Spec{Runtime: p.name}
	for _, field := range p.fields {
		values, err := field.Values()
		if err != nil {
			return nil, err
		}

		if spec.Metadata == nil {
			spec.Metadata = map[string]*adagio.MetadataValue{}
		}

		spec.Metadata[p.fieldName(field)] = &adagio.MetadataValue{Values: values}
	}

	return spec, nil
}

// Parse parses the state from a node spec into the targets
// set on the builders fields
func (p *Builder) Parse(spec *adagio.Node_Spec) error {
	for _, field := range p.fields {
		values, ok := spec.Metadata[p.fieldName(field)]
		if !ok {
			if field.Required {
				return errors.New("key not found")
			}

			if err := field.SetToDefault(); err != nil {
				return err
			}

			continue
		}

		if err := field.Parse(values.Values); err != nil {
			return err
		}
	}

	return nil
}

// StringField is a function which takes a string pointer
// and returns a Field pointers which will set and get from
// the provided string pointer
type StringField func(*string) *Field

// String configures a StringField which when call with a string pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.Spec()
func (s *Builder) String(name string, required bool, defaultValue string) StringField {
	field := newField(name, StringFieldType, required, []string{defaultValue})
	s.fields = append(s.fields, field)

	return func(v *string) *Field {
		field.values = func() ([]string, error) {
			if v == nil {
				if required {
					return nil, errors.New("field is required")
				}

				return nil, nil
			}

			return []string{*v}, nil
		}

		field.parse = func(vs []string) error {
			if len(vs) < 1 {
				return errors.New("no value set for key")
			}

			*v = vs[0]

			return nil
		}

		return field
	}
}

// StringsField is a function which takes a slice of string pointer
// and returns a Field pointers which will set and get from
// the provided slice of string pointer
type StringsField func(*[]string) *Field

// Strings configures a StringsField which when call with a string slice pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.Spec()
func (s *Builder) Strings(name string, required bool, defaultValues ...string) StringsField {
	field := newField(name, StringsFieldType, required, defaultValues)
	s.fields = append(s.fields, field)

	return func(v *[]string) *Field {
		field.values = func() ([]string, error) {
			if v == nil {
				if required {
					return nil, errors.New("field is required")
				}

				return nil, nil
			}

			return *v, nil
		}

		field.parse = func(vs []string) error {
			*v = vs

			return nil
		}

		return field
	}
}

// Int64Field is a function which takes a int64 pointer
// and returns a Field pointers which will set and get from
// the provided pointer
type Int64Field func(*int64) *Field

// Int64 configures a Int64Field which when call with a int64 pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.Spec()
func (s *Builder) Int64(name string, required bool, defaultValue int64) Int64Field {
	field := newField(name, Int64FieldType, required, []string{fmt.Sprintf("%d", defaultValue)})
	s.fields = append(s.fields, field)

	return func(v *int64) *Field {
		field.values = func() ([]string, error) {
			if v == nil {
				if required {
					return nil, errors.New("field is required")
				}

				return nil, nil
			}

			return []string{fmt.Sprintf("%d", *v)}, nil
		}

		field.parse = func(vs []string) (err error) {
			if len(vs) < 1 {
				return errors.New("no value set for key")
			}

			*v, err = strconv.ParseInt(vs[0], 10, 64)

			return err
		}

		return field
	}
}

// TimeField is a function which takes a int64 pointer
// and returns a Field pointers which will set and get from
// the provided pointer
type TimeField func(*time.Time) *Field

// Time configures a TimeField which when call with a int64 pointer
// will set the pointer on calls to builder.Parse() and read the value
// at the end of the pointer on calls to builder.Spec()
func (s *Builder) Time(name string, required bool, defaultValue time.Time) TimeField {
	field := newField(name, TimeFieldType, required, []string{defaultValue.Format(time.RFC3339Nano)})
	s.fields = append(s.fields, field)

	return func(v *time.Time) *Field {
		field.values = func() ([]string, error) {
			if v == nil {
				if required {
					return nil, errors.New("field is required")
				}

				return nil, nil
			}

			return []string{v.Format(time.RFC3339Nano)}, nil
		}

		field.parse = func(vs []string) (err error) {
			if len(vs) < 1 {
				return errors.New("no value set for key")
			}

			*v, err = time.Parse(time.RFC3339Nano, vs[0])

			return err
		}

		return field
	}
}

// Field is a structure which contains the properties
// of a field used to call a runtime
type Field struct {
	Name     string
	Type     FieldType
	Required bool
	Defaults []string
	parse    func([]string) error
	values   func() ([]string, error)
}

// Values returns the fields value represented as a slice of strings
// for use in node spec metadata
func (f *Field) Values() ([]string, error) {
	if f.values == nil {
		return nil, fmt.Errorf("target not set for %q", f.Name)
	}

	return f.values()
}

// Parse sets the values of the field using the provided slice of strings
func (f *Field) Parse(args []string) error {
	if f.parse == nil {
		return fmt.Errorf("target not set for %q", f.Name)
	}

	return f.parse(args)
}

// SetToDefault applies the default values to the field
func (f Field) SetToDefault() error {
	return f.Parse(f.Defaults)
}

func newField(name string, fieldType FieldType, required bool, defaultValues []string) *Field {
	return &Field{
		Name:     name,
		Type:     fieldType,
		Required: required,
		Defaults: defaultValues,
	}
}
