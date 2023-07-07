package run

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	compout "github.com/powertoolsdev/mono/pkg/types/plugins/component/v1"
	"github.com/stretchr/testify/assert"
)

func Test_run_getApplyPipeline(t *testing.T) {
	t.Run("test that apply pipeline is correct", func(t *testing.T) {
		//
	})
}

func raw(name string) json.RawMessage {
	rm := json.RawMessage{}
	_ = rm.UnmarshalJSON([]byte(name))
	return rm
}

func TestTerraformToNuon(t *testing.T) {
	tests := map[string]struct {
		in       map[string]tfexec.OutputMeta
		expected *compout.Outputs
	}{
		"empty case": {
			in:       map[string]tfexec.OutputMeta{},
			expected: &compout.Outputs{},
		},
		"boolean case": {
			in: map[string]tfexec.OutputMeta{
				"enableCache": {Sensitive: true, Type: raw("boolean"), Value: raw("true")},
			},
			expected: &compout.Outputs{
				Values: []*compout.Value{{
					Scalar:    &compout.Value_Bool{Bool: true},
					Path:      []string{"enableCache"},
					Sensitive: true,
				}},
			},
		},
		"number case": {
			in: map[string]tfexec.OutputMeta{
				"cacheSize": {Sensitive: true, Type: raw("number"), Value: raw("42.42")},
			},
			expected: &compout.Outputs{
				Values: []*compout.Value{{
					Scalar:    &compout.Value_Double{Double: 42.42},
					Path:      []string{"cacheSize"},
					Sensitive: true,
				}},
			},
		},
		"string case": {
			in: map[string]tfexec.OutputMeta{
				"apiKey": {Sensitive: true, Type: raw("string"), Value: raw(`"ak_123456"`)},
			},
			expected: &compout.Outputs{
				Values: []*compout.Value{{
					Scalar:    &compout.Value_String_{String_: "ak_123456"},
					Path:      []string{"apiKey"},
					Sensitive: true,
				}},
			},
		},
		"object case": {
			in: map[string]tfexec.OutputMeta{
				"server": {Sensitive: true, Type: raw("number"), Value: raw(`{"barkey":true,"fookey":"FOOVALUE"}`)},
			},
			expected: &compout.Outputs{
				Values: []*compout.Value{
					{
						Path:      []string{"server"},
						Sensitive: true,
					},
					{
						Scalar: &compout.Value_Bool{Bool: true},
						Path:   []string{"server", "barkey"},
					},
					{
						Scalar: &compout.Value_String_{String_: "FOOVALUE"},
						Path:   []string{"server", "fookey"},
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := terraformToNuon(test.in)
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
