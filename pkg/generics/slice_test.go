package generics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceToGroups(t *testing.T) {
	tests := map[string]struct {
		input  []string
		output [][]string
		size   int
	}{
		"single item with no limit": {
			input:  []string{"a"},
			output: [][]string{{"a"}},
			size:   -1,
		},
		"single item with limit 1": {
			input:  []string{"a"},
			output: [][]string{{"a"}},
			size:   1,
		},
		"single item with limit": {
			input:  []string{"a"},
			output: [][]string{{"a"}},
			size:   10000,
		},

		"multiple items with limit of 1": {
			input:  []string{"a", "b", "c"},
			output: [][]string{{"a"}, {"b"}, {"c"}},
			size:   1,
		},
		"multiple items with limit of -1": {
			input:  []string{"a", "b", "c"},
			output: [][]string{{"a", "b", "c"}},
			size:   -1,
		},
		"multiple items with limit": {
			input:  []string{"a", "b", "c"},
			output: [][]string{{"a", "b"}, {"c"}},
			size:   2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output := SliceToGroups(test.input, test.size)
			require.Equal(t, test.output, output)
		})
	}
}
