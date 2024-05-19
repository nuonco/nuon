package generics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirst(t *testing.T) {
	tests := map[string]struct {
		fn       func() interface{}
		expected interface{}
	}{
		"strings": {
			fn: func() interface{} {
				return First("", "a")
			},
			expected: "a",
		},
		"bools": {
			fn: func() interface{} {
				return First(false, true)
			},
			expected: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res := tt.fn()
			require.Equal(t, tt.expected, res)
		})
	}
}
