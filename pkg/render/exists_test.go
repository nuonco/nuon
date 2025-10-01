package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	testCases := map[string]struct {
		input  map[string]any
		lookup string
		value  bool
	}{
		"ok": {
			input: map[string]any{
				"nuon": map[string]any{
					"key": "value",
				},
			},
			lookup: "nuon.key",
			value:  true,
		},
		"still-ok-with-dot": {
			input: map[string]any{
				"nuon": map[string]any{
					"key": "value",
				},
			},
			lookup: ".nuon.key",
			value:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			output := Exists(tc.lookup, tc.input)
			require.Equal(t, tc.value, output)
		})
	}
}
