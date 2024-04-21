package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	type args struct {
		inputCfg  string
		outputCfg string
	}
	tests := []struct {
		name      string
		inputCfg  string
		outputCfg string
		err       error
	}{
		{
			name: "base case with vars",
			inputCfg: `version = "v1"
[config_vars]
key = "value"

[[components.val]]
key = "{{.key}}"
			`,
			outputCfg: `version = "v1"
[config_vars]
key = "value"

[[components.val]]
key = "value"
			`,
		},
		{
			name: "base case no vars",
			inputCfg: `version = "v1"
[[components.val]]
key = "{{.key}}"
			`,
			outputCfg: `version = "v1"
[[components.val]]
key = "{{.key}}"
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderedCfg, err := Template([]byte(tt.inputCfg))

			require.Nil(t, err)
			require.Equal(t, tt.err == nil, err == nil, "did not match error/non-error expectation")
			require.Equal(t, tt.outputCfg, string(renderedCfg))
		})
	}
}
