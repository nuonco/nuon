package render

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RenderMapTestSuite is the testify suite for render_map tests
type RenderMapTestSuite struct {
	suite.Suite
}

// TestRenderMapSuite runs the test suite
func TestRenderMapSuite(t *testing.T) {
	suite.Run(t, new(RenderMapTestSuite))
}

type renderMapTest struct {
	input       map[string]string
	data        map[string]any
	expected    any
	shouldError bool
}

func (s *RenderMapTestSuite) TestRenderMapWithStringMap() {
	t := s.T()
	tests := map[string]renderMapTest{
		"simple map[string]string": {
			input: map[string]string{
				"id":      "{{.nuon.install.id}}",
				"static":  "hello",
				"another": "install-{{.nuon.install.id}}-suffix",
			},
			data: map[string]any{
				"nuon": map[string]any{
					"install": map[string]any{
						"id": "abc123",
					},
				},
			},
			expected: map[string]string{
				"id":      "abc123",
				"static":  "hello",
				"another": "install-abc123-suffix",
			},
			shouldError: false,
		},
		"missing nuon value": {
			input: map[string]string{
				"key": "{{.nuon.install.id}}",
			},
			data:        map[string]any{},
			expected:    nil,
			shouldError: true,
		},
		"missing non nuon value": {
			input: map[string]string{
				"key": "{{.foo.install.id}}",
			},
			expected: map[string]string{
				"key": "{{.foo.install.id}}",
			},
			data:        map[string]any{},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderMap(&tc.input, tc.data)
			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

type renderMapAnyTest struct {
	input       map[string]any
	data        map[string]any
	expected    any
	shouldError bool
}

func (s *RenderMapTestSuite) TestRenderMapWithStringMapAny() {
	t := s.T()
	tests := map[string]renderMapAnyTest{
		"simple map[string]string": {
			input: map[string]any{
				"id":      "{{.nuon.install.id}}",
				"static":  "hello",
				"another": "install-{{.nuon.install.id}}-suffix",
			},
			data: map[string]any{
				"nuon": map[string]any{
					"install": map[string]any{
						"id": "abc123",
					},
				},
			},
			expected: map[string]any{
				"id":      "abc123",
				"static":  "hello",
				"another": "install-abc123-suffix",
			},
			shouldError: false,
		},
		"missing nuon value": {
			input: map[string]any{
				"key": "{{.nuon.install.id}}",
			},
			data:        map[string]any{},
			expected:    nil,
			shouldError: true,
		},
		"missing non nuon value": {
			input: map[string]any{
				"key": "{{.foo.install.id}}",
			},
			expected: map[string]any{
				"key": "{{.foo.install.id}}",
			},
			data:        map[string]any{},
			shouldError: false,
		},
		"nested": {
			input: map[string]any{
				"key": "{{.nuon.install.id}}",
				"nested": map[string]string{
					"key": "{{.nuon.install.id}}",
				},
			},
			expected: map[string]any{
				"key": "abc123",
				"nested": map[string]string{
					"key": "abc123",
				},
			},
			data: map[string]any{
				"nuon": map[string]any{
					"install": map[string]any{
						"id": "abc123",
					},
				},
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderMap(&tc.input, tc.data)
			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}
