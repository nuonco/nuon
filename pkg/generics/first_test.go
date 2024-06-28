package generics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirst(t *testing.T) {
	var emptystr string
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
		"ptrs": {
			// Empty string value is returned because First relies on simple equality checks
			// that compare pointer addresses
			fn: func() any {
				return First(&emptystr, ToPtr("foo"))
			},
			expected: &emptystr,
		},
		"nilptrs": {
			// But a nil pointer in the zero index is skipped
			fn: func() any {
				return First(nil, &emptystr)
			},
			expected: &emptystr,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res := tt.fn()
			require.Equal(t, tt.expected, res)
			if res != tt.expected {
				t.Fatal("require.Equal succeeds but basic equality check fails")
			}
		})
	}
}
