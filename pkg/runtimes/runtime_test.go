package runtime

import (
	"testing"
	"time"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/stretchr/testify/assert"
)

var defaultTime = time.Date(2019, 7, 10, 10, 0, 0, 50, time.UTC)

func Test_Builder_Spec(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		setup func() *Builder
		spec  *adagio.Node_Spec
	}{
		{
			name: "happy path",
			setup: func() *Builder {
				var (
					builder      = NewBuilder("foo")
					stringField  = "a_string"
					stringsField = []string{"a", "b", "c"}
					int64Field   = int64(12345)
					timeField    = time.Date(2019, 1, 1, 10, 0, 0, 75, time.UTC)
				)

				builder.String(&stringField, "string_field", false, "")
				builder.Strings(&stringsField, "strings_field", false)
				builder.Int64(&int64Field, "int64_field", false, 0)
				builder.Time(&timeField, "time_field", false, defaultTime)

				return builder
			},
			spec: &adagio.Node_Spec{
				Name:    "happy path",
				Runtime: "foo",
				Metadata: map[string]*adagio.MetadataValue{
					"adagio.arguments.foo.string_field": &adagio.MetadataValue{
						Values: []string{"a_string"},
					},
					"adagio.arguments.foo.strings_field": &adagio.MetadataValue{
						Values: []string{"a", "b", "c"},
					},
					"adagio.arguments.foo.int64_field": &adagio.MetadataValue{
						Values: []string{"12345"},
					},
					"adagio.arguments.foo.time_field": &adagio.MetadataValue{
						Values: []string{"2019-01-01T10:00:00.000000075Z"},
					},
				},
			},
		},
		{
			name: "happy path - set argument from input",
			setup: func() *Builder {
				var (
					builder      = NewBuilder("foo")
					stringField  = "a_string"
					stringsField = []string{"a", "b", "c"}
					int64Field   = int64(12345)
					timeField    = time.Date(2019, 1, 1, 10, 0, 0, 75, time.UTC)
				)

				builder.String(&stringField, "string_field", false, "")
				builder.Strings(&stringsField, "strings_field", false)
				builder.Int64(&int64Field, "int64_field", false, 0)
				builder.Time(&timeField, "time_field", false, defaultTime)

				if err := builder.SetArgumentFromInput("int64_field", "other_func"); err != nil {
					t.Fatal(err)
				}

				return builder
			},
			spec: &adagio.Node_Spec{
				Name:    "happy path - set argument from input",
				Runtime: "foo",
				Metadata: map[string]*adagio.MetadataValue{
					"adagio.arguments.foo.string_field": &adagio.MetadataValue{
						Values: []string{"a_string"},
					},
					"adagio.arguments.foo.strings_field": &adagio.MetadataValue{
						Values: []string{"a", "b", "c"},
					},
					"adagio.arguments.foo.time_field": &adagio.MetadataValue{
						Values: []string{"2019-01-01T10:00:00.000000075Z"},
					},
					"adagio.inputs.foo.int64_field": &adagio.MetadataValue{
						Values: []string{"other_func"},
					},
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			b := testCase.setup()

			spec, err := b.NewSpec(testCase.name)
			assert.Nil(t, err)

			assert.Equal(t, testCase.spec, spec)
		})
	}
}

func Test_Builder_Parse(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		node  *adagio.Node
		setup func() (*Builder, func(*testing.T))
	}{
		{
			name: "happy path",
			node: &adagio.Node{
				Spec: &adagio.Node_Spec{
					Name:    "a",
					Runtime: "foo",
					Metadata: map[string]*adagio.MetadataValue{
						"adagio.arguments.foo.string_field": &adagio.MetadataValue{
							Values: []string{"a_string"},
						},
						"adagio.arguments.foo.strings_field": &adagio.MetadataValue{
							Values: []string{"a", "b", "c"},
						},
						"adagio.arguments.foo.int64_field": &adagio.MetadataValue{
							Values: []string{"12345"},
						},
						"adagio.arguments.foo.time_field": &adagio.MetadataValue{
							Values: []string{"2019-07-10T10:00:00.000000050Z"},
						},
					},
				},
			},
			setup: func() (*Builder, func(*testing.T)) {
				var (
					builder      = NewBuilder("foo")
					stringField  string
					stringsField []string
					int64Field   int64
					timeField    time.Time
				)

				builder.String(&stringField, "string_field", false, "")
				builder.Strings(&stringsField, "strings_field", false)
				builder.Int64(&int64Field, "int64_field", false, 0)
				builder.Time(&timeField, "time_field", false, defaultTime)

				return builder, func(t *testing.T) {
					assert.Equal(t, "a_string", stringField)
					assert.Equal(t, []string{"a", "b", "c"}, stringsField)
					assert.Equal(t, int64(12345), int64Field)
					assert.Equal(t, defaultTime, timeField)
				}
			},
		},
		{
			name: "happy path - defaults",
			node: &adagio.Node{
				Spec: &adagio.Node_Spec{
					Name:     "a",
					Runtime:  "foo",
					Metadata: map[string]*adagio.MetadataValue{},
				},
			},
			setup: func() (*Builder, func(*testing.T)) {
				var (
					builder      = NewBuilder("foo")
					stringField  string
					stringsField []string
					int64Field   int64
				)

				builder.String(&stringField, "string_field", false, "other_string")
				builder.Strings(&stringsField, "strings_field", false, "c", "d", "e")
				builder.Int64(&int64Field, "int64_field", false, 20)

				return builder, func(t *testing.T) {
					assert.Equal(t, "other_string", stringField)
					assert.Equal(t, []string{"c", "d", "e"}, stringsField)
					assert.Equal(t, int64(20), int64Field)
				}
			},
		},
		{
			name: "happy path - from inputs",
			node: &adagio.Node{
				Spec: &adagio.Node_Spec{
					Name:    "a",
					Runtime: "foo",
					Metadata: map[string]*adagio.MetadataValue{
						"adagio.arguments.foo.string_field": &adagio.MetadataValue{
							Values: []string{"a_string"},
						},
						"adagio.arguments.foo.strings_field": &adagio.MetadataValue{
							Values: []string{"a", "b", "c"},
						},
						"adagio.arguments.foo.int64_field": &adagio.MetadataValue{
							Values: []string{"12345"},
						},
						"adagio.inputs.foo.time_field": &adagio.MetadataValue{
							Values: []string{"other_func"},
						},
					},
				},
				Inputs: map[string][]byte{
					"other_func": []byte("2019-01-01T10:00:00.000000075Z"),
				},
			},
			setup: func() (*Builder, func(*testing.T)) {
				var (
					builder      = NewBuilder("foo")
					stringField  string
					stringsField []string
					int64Field   int64
					timeField    time.Time
				)

				builder.String(&stringField, "string_field", false, "")
				builder.Strings(&stringsField, "strings_field", false)
				builder.Int64(&int64Field, "int64_field", false, 0)
				builder.Time(&timeField, "time_field", false, defaultTime)

				return builder, func(t *testing.T) {
					assert.Equal(t, "a_string", stringField)
					assert.Equal(t, []string{"a", "b", "c"}, stringsField)
					assert.Equal(t, int64(12345), int64Field)
					assert.Equal(t, time.Date(2019, 1, 1, 10, 0, 0, 75, time.UTC), timeField)
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			builder, assert := testCase.setup()

			builder.Parse(testCase.node)

			assert(t)
		})
	}
}
