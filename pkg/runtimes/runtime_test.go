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

				builder.String("string_field", false, "")(&stringField)
				builder.Strings("strings_field", false)(&stringsField)
				builder.Int64("int64_field", false, 0)(&int64Field)
				builder.Time("time_field", false, defaultTime)(&timeField)

				return builder
			},
			spec: &adagio.Node_Spec{
				Runtime: "foo",
				Metadata: map[string]*adagio.MetadataValue{
					"adagio.runtime.foo.string_field": &adagio.MetadataValue{
						Values: []string{"a_string"},
					},
					"adagio.runtime.foo.strings_field": &adagio.MetadataValue{
						Values: []string{"a", "b", "c"},
					},
					"adagio.runtime.foo.int64_field": &adagio.MetadataValue{
						Values: []string{"12345"},
					},
					"adagio.runtime.foo.time_field": &adagio.MetadataValue{
						Values: []string{"2019-01-01T10:00:00.000000075Z"},
					},
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			b := testCase.setup()

			spec, err := b.NewSpec()
			assert.Nil(t, err)

			assert.Equal(t, testCase.spec, spec)
		})
	}
}

func Test_Builder_Parse(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		spec  *adagio.Node_Spec
		setup func() (*Builder, func(*testing.T))
	}{
		{
			name: "happy path",
			spec: &adagio.Node_Spec{
				Name:    "a",
				Runtime: "foo",
				Metadata: map[string]*adagio.MetadataValue{
					"adagio.runtime.foo.string_field": &adagio.MetadataValue{
						Values: []string{"a_string"},
					},
					"adagio.runtime.foo.strings_field": &adagio.MetadataValue{
						Values: []string{"a", "b", "c"},
					},
					"adagio.runtime.foo.int64_field": &adagio.MetadataValue{
						Values: []string{"12345"},
					},
					"adagio.runtime.foo.time_field": &adagio.MetadataValue{
						Values: []string{"2019-07-10T10:00:00.000000050Z"},
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

				builder.String("string_field", false, "")(&stringField)
				builder.Strings("strings_field", false)(&stringsField)
				builder.Int64("int64_field", false, 0)(&int64Field)
				builder.Time("time_field", false, defaultTime)(&timeField)

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
			spec: &adagio.Node_Spec{
				Name:     "a",
				Runtime:  "foo",
				Metadata: map[string]*adagio.MetadataValue{},
			},
			setup: func() (*Builder, func(*testing.T)) {
				var (
					builder      = NewBuilder("foo")
					stringField  string
					stringsField []string
					int64Field   int64
				)

				builder.String("string_field", false, "other_string")(&stringField)
				builder.Strings("strings_field", false, "c", "d", "e")(&stringsField)
				builder.Int64("int64_field", false, 20)(&int64Field)

				return builder, func(t *testing.T) {
					assert.Equal(t, "other_string", stringField)
					assert.Equal(t, []string{"c", "d", "e"}, stringsField)
					assert.Equal(t, int64(20), int64Field)
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			builder, assert := testCase.setup()

			builder.Parse(&adagio.Node{Spec: testCase.spec})

			assert(t)
		})
	}
}
