package outputs

import (
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestTFOutputMetaToStructPB(t *testing.T) {
	tests := map[string]struct {
		in       map[string]tfexec.OutputMeta
		expected func() (*structpb.Struct, error)
	}{
		"empty case": {
			in: map[string]tfexec.OutputMeta{},
			expected: func() (*structpb.Struct, error) {
				return structpb.NewStruct(nil)
			},
		},
		"boolean case": {
			in: map[string]tfexec.OutputMeta{
				"enableCache": {Sensitive: true, Type: []byte("boolean"), Value: []byte("true")},
			},
			expected: func() (*structpb.Struct, error) {
				return structpb.NewStruct(map[string]interface{}{"enableCache": true})
			},
		},
		"number case": {
			in: map[string]tfexec.OutputMeta{
				"cacheSize": {Sensitive: true, Type: []byte("number"), Value: []byte("42.42")},
			},
			expected: func() (*structpb.Struct, error) {
				return structpb.NewStruct(map[string]interface{}{"cacheSize": 42.42})
			},
		},
		"string case": {
			in: map[string]tfexec.OutputMeta{
				"apiKey": {Sensitive: true, Type: []byte("string"), Value: []byte(`"ak_123456"`)},
			},
			expected: func() (*structpb.Struct, error) {
				return structpb.NewStruct(map[string]interface{}{"apiKey": "ak_123456"})
			},
		},
		"object case": {
			in: map[string]tfexec.OutputMeta{
				"server": {Sensitive: true, Type: []byte("object"), Value: []byte(`{"barkey":true,"fookey":"FOOVALUE"}`)},
			},
			expected: func() (*structpb.Struct, error) {
				innerMap := map[string]interface{}{"barkey": true, "fookey": "FOOVALUE"}
				outerMap := map[string]interface{}{"server": innerMap}
				return structpb.NewStruct(outerMap)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := TFOutputMetaToStructPB(test.in)
			assert.NoError(t, err)
			expected, err := test.expected()
			assert.NoError(t, err)
			assert.EqualValues(t, expected.AsMap(), actual.AsMap())
		})
	}
}
