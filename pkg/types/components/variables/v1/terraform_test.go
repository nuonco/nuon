package variablesv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformVariables_ToMap(t *testing.T) {
	tests := map[string]struct {
		vars     *TerraformVariables
		expected map[string]string
	}{
		"no conflicts": {
			vars: &TerraformVariables{
				Variables: []*TerraformVariable{
					{
						Name:  "key",
						Value: "value",
					},
				},
			},
			expected: map[string]string{
				"key": "value",
			},
		},
		"conflicts uses last value": {
			vars: &TerraformVariables{
				Variables: []*TerraformVariable{
					{
						Name:  "key",
						Value: "value",
					},
					{
						Name:  "key",
						Value: "value-2",
					},
				},
			},
			expected: map[string]string{
				"key": "value-2",
			},
		},
		"handles nil": {
			vars:     nil,
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res := test.vars.ToMap()
			assert.Equal(t, test.expected, res)
		})
	}
}
