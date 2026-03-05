package render

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ExistsTestSuite is the testify suite for exists tests
type ExistsTestSuite struct {
	suite.Suite
}

// TestExistsSuite runs the test suite
func TestExistsSuite(t *testing.T) {
	suite.Run(t, new(ExistsTestSuite))
}

func (s *ExistsTestSuite) TestExists() {
	t := s.T()
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
