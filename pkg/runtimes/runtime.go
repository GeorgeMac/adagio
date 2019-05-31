package runtime

import (
	"errors"
	"fmt"

	"github.com/georgemac/adagio/pkg/adagio"
)

func Parse(spec *adagio.Node_Spec, fields ...Field) error {
	for _, field := range fields {
		name := fmt.Sprintf("adagio.runtime.%s", field.Name())

		values, ok := spec.Metadata[name]
		if !ok {
			if field.Required() {
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

type Spec struct {
	Name string
}

func NewSpec(name string) *Spec {
	return &Spec{name}
}

func (s *Spec) String(name string, required bool, defaultValue string) StringField {
	return func(v *string) Field {
		return newField(fmt.Sprintf("%s.%s", s.Name, name), required, []string{defaultValue}, func(vs []string) error {
			if len(vs) < 1 {
				return errors.New("no value set for key")
			}

			*v = vs[0]

			return nil
		})
	}
}

func (s *Spec) Strings(name string, required bool, defaultValue []string) StringsField {
	return func(v *[]string) Field {
		return newField(fmt.Sprintf("%s.%s", s.Name, name), required, defaultValue, func(vs []string) error {
			*v = vs

			return nil
		})
	}
}

type StringField func(*string) Field

type StringsField func(*[]string) Field

type Field struct {
	name         string
	required     bool
	parse        func([]string) error
	setToDefault func() error
}

func (f Field) Name() string { return f.name }

func (f Field) Required() bool { return f.required }

func (f Field) Parse(args []string) error { return f.parse(args) }

func (f Field) SetToDefault() error { return f.setToDefault() }

func newField(name string, required bool, defaultValues []string, parse func([]string) error) Field {
	return Field{
		name:     name,
		required: required,
		parse:    parse,
		setToDefault: func() error {
			return parse(defaultValues)
		},
	}
}
