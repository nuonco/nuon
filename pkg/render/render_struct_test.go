package render

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test fixture structs
type simpleStruct struct {
	ID         string `features:"template"`
	Name       string `features:"template"`
	NoTemplate string
}

type childStruct struct {
	Value string `features:"template"`
}

type nestedStruct struct {
	Parent string `features:"template"`
	Child  *childStruct
}

type deeplyNestedStruct struct {
	Level1 string `features:"template"`
	Nested *level2Struct
}

type level2Struct struct {
	Level2 string `features:"template"`
	Nested *level3Struct
}

type level3Struct struct {
	Level3 string `features:"template"`
}

type itemConfig struct {
	Name string `features:"template"`
}

type sliceStruct struct {
	Items   []string `features:"template"`
	Configs []*itemConfig
}

type mapStruct struct {
	Labels   map[string]string `features:"template"`
	Metadata map[string]any    `features:"template"`
}

type nestedConfig struct {
	Namespace string `features:"template"`
	Manifest  string `features:"template"`
}

type complexStruct struct {
	Name   string `features:"template"`
	Config *nestedConfig
	Items  []*itemConfig
	Labels map[string]string `features:"template"`
	Data   []byte            `features:"template"`
}

type realWorldConfig struct {
	Manifest  string `features:"template"`
	Namespace string `features:"template"`
	Kustomize *kustomizeConfig
}

type kustomizeConfig struct {
	Path string `features:"template"`
}

type unexportedFieldStruct struct {
	Public  string `features:"template"`
	private string `features:"template"`
}

// RenderStructTestSuite is the testify suite for render_struct tests
type RenderStructTestSuite struct {
	suite.Suite
}

// TestRenderStructSuite runs the test suite
func TestRenderStructSuite(t *testing.T) {
	suite.Run(t, new(RenderStructTestSuite))
}

// Test data sets
func getStandardData() map[string]any {
	return map[string]any{
		"nuon": map[string]any{
			"install": map[string]any{
				"id":        "inst-abc123",
				"name":      "production",
				"namespace": "prod-ns",
			},
			"app": map[string]any{
				"id":   "app-xyz789",
				"name": "myapp",
			},
			"org": map[string]any{
				"id":   "org-123456",
				"name": "acme-corp",
			},
		},
	}
}

func getMinimalData() map[string]any {
	return map[string]any{
		"nuon": map[string]any{
			"install": map[string]any{
				"id": "inst-minimal",
			},
		},
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_SimpleStringFields() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
		errorMsg    string
	}{
		"single template field": {
			input: &simpleStruct{
				ID: "{{.nuon.install.id}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "inst-abc123",
			},
			shouldError: false,
		},
		"multiple template fields": {
			input: &simpleStruct{
				ID:   "{{.nuon.install.id}}",
				Name: "{{.nuon.install.name}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID:   "inst-abc123",
				Name: "production",
			},
			shouldError: false,
		},
		"field without template tag skipped": {
			input: &simpleStruct{
				ID:         "{{.nuon.install.id}}",
				NoTemplate: "{{.nuon.install.name}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID:         "inst-abc123",
				NoTemplate: "{{.nuon.install.name}}",
			},
			shouldError: false,
		},
		"empty string field": {
			input: &simpleStruct{
				ID:   "",
				Name: "{{.nuon.install.name}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID:   "",
				Name: "production",
			},
			shouldError: false,
		},
		"missing data error": {
			input: &simpleStruct{
				ID: "{{.nuon.missing.value}}",
			},
			data:        getMinimalData(),
			shouldError: true,
		},
		"non-nuon template preserved": {
			input: &simpleStruct{
				ID: "{{.other.value}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "{{.other.value}}",
			},
			shouldError: false,
		},
		"template with static text": {
			input: &simpleStruct{
				ID: "prefix-{{.nuon.install.id}}-suffix",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "prefix-inst-abc123-suffix",
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_NestedStructs() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
	}{
		"single level nesting": {
			input: &nestedStruct{
				Parent: "{{.nuon.install.id}}",
				Child: &childStruct{
					Value: "{{.nuon.install.name}}",
				},
			},
			data: getStandardData(),
			expected: &nestedStruct{
				Parent: "inst-abc123",
				Child: &childStruct{
					Value: "production",
				},
			},
			shouldError: false,
		},
		"three level deep nesting": {
			input: &deeplyNestedStruct{
				Level1: "{{.nuon.install.id}}",
				Nested: &level2Struct{
					Level2: "{{.nuon.app.id}}",
					Nested: &level3Struct{
						Level3: "{{.nuon.org.id}}",
					},
				},
			},
			data: getStandardData(),
			expected: &deeplyNestedStruct{
				Level1: "inst-abc123",
				Nested: &level2Struct{
					Level2: "app-xyz789",
					Nested: &level3Struct{
						Level3: "org-123456",
					},
				},
			},
			shouldError: false,
		},
		"nil pointer skipped": {
			input: &nestedStruct{
				Parent: "{{.nuon.install.id}}",
				Child:  nil,
			},
			data: getStandardData(),
			expected: &nestedStruct{
				Parent: "inst-abc123",
				Child:  nil,
			},
			shouldError: false,
		},
		"partial nesting with nil": {
			input: &deeplyNestedStruct{
				Level1: "{{.nuon.install.id}}",
				Nested: &level2Struct{
					Level2: "{{.nuon.app.id}}",
					Nested: nil,
				},
			},
			data: getStandardData(),
			expected: &deeplyNestedStruct{
				Level1: "inst-abc123",
				Nested: &level2Struct{
					Level2: "app-xyz789",
					Nested: nil,
				},
			},
			shouldError: false,
		},
		"parent and child with templates": {
			input: &nestedStruct{
				Parent: "parent-{{.nuon.install.id}}",
				Child: &childStruct{
					Value: "child-{{.nuon.install.id}}",
				},
			},
			data: getStandardData(),
			expected: &nestedStruct{
				Parent: "parent-inst-abc123",
				Child: &childStruct{
					Value: "child-inst-abc123",
				},
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_Slices() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
	}{
		"slice of strings with templates": {
			input: &sliceStruct{
				Items: []string{
					"{{.nuon.install.id}}",
					"{{.nuon.app.id}}",
					"static-value",
				},
			},
			data: getStandardData(),
			expected: &sliceStruct{
				Items: []string{
					"inst-abc123",
					"app-xyz789",
					"static-value",
				},
			},
			shouldError: false,
		},
		"slice of pointer structs": {
			input: &sliceStruct{
				Configs: []*itemConfig{
					{Name: "{{.nuon.install.id}}"},
					{Name: "{{.nuon.app.id}}"},
				},
			},
			data: getStandardData(),
			expected: &sliceStruct{
				Configs: []*itemConfig{
					{Name: "inst-abc123"},
					{Name: "app-xyz789"},
				},
			},
			shouldError: false,
		},
		"slice with nil elements": {
			input: &sliceStruct{
				Configs: []*itemConfig{
					{Name: "{{.nuon.install.id}}"},
					nil,
					{Name: "{{.nuon.app.id}}"},
				},
			},
			data: getStandardData(),
			expected: &sliceStruct{
				Configs: []*itemConfig{
					{Name: "inst-abc123"},
					nil,
					{Name: "app-xyz789"},
				},
			},
			shouldError: false,
		},
		"empty slice": {
			input: &sliceStruct{
				Items: []string{},
			},
			data: getStandardData(),
			expected: &sliceStruct{
				Items: []string{},
			},
			shouldError: false,
		},
		"byte slice with template": {
			input: &complexStruct{
				Data: []byte("{{.nuon.install.id}}"),
			},
			data: getStandardData(),
			expected: &complexStruct{
				Data: []byte("inst-abc123"),
			},
			shouldError: false,
		},
		"slice with mixed templates and static": {
			input: &sliceStruct{
				Items: []string{
					"prefix-{{.nuon.install.id}}",
					"static",
					"{{.nuon.app.id}}-suffix",
				},
			},
			data: getStandardData(),
			expected: &sliceStruct{
				Items: []string{
					"prefix-inst-abc123",
					"static",
					"app-xyz789-suffix",
				},
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_Maps() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
	}{
		"map with template tag": {
			input: &mapStruct{
				Labels: map[string]string{
					"id":   "{{.nuon.install.id}}",
					"name": "{{.nuon.install.name}}",
				},
			},
			data: getStandardData(),
			expected: &mapStruct{
				Labels: map[string]string{
					"id":   "inst-abc123",
					"name": "production",
				},
			},
			shouldError: false,
		},
		"map[string]any with string values": {
			input: &mapStruct{
				Metadata: map[string]any{
					"id":     "{{.nuon.install.id}}",
					"name":   "{{.nuon.install.name}}",
					"static": "value",
				},
			},
			data: getStandardData(),
			expected: &mapStruct{
				Metadata: map[string]any{
					"id":     "inst-abc123",
					"name":   "production",
					"static": "value",
				},
			},
			shouldError: false,
		},
		"empty map": {
			input: &mapStruct{
				Labels: map[string]string{},
			},
			data: getStandardData(),
			expected: &mapStruct{
				Labels: map[string]string{},
			},
			shouldError: false,
		},
		"map with nested templates": {
			input: &mapStruct{
				Labels: map[string]string{
					"key": "{{.nuon.install.id}}-{{.nuon.app.id}}",
				},
			},
			data: getStandardData(),
			expected: &mapStruct{
				Labels: map[string]string{
					"key": "inst-abc123-app-xyz789",
				},
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_ComplexScenarios() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
	}{
		"deeply nested with slices and maps": {
			input: &complexStruct{
				Name: "{{.nuon.install.id}}",
				Config: &nestedConfig{
					Namespace: "{{.nuon.install.namespace}}",
					Manifest:  "{{.nuon.app.name}}",
				},
				Items: []*itemConfig{
					{Name: "{{.nuon.install.id}}"},
					{Name: "{{.nuon.app.id}}"},
				},
				Labels: map[string]string{
					"install": "{{.nuon.install.id}}",
					"app":     "{{.nuon.app.id}}",
				},
				Data: []byte("{{.nuon.org.name}}"),
			},
			data: getStandardData(),
			expected: &complexStruct{
				Name: "inst-abc123",
				Config: &nestedConfig{
					Namespace: "prod-ns",
					Manifest:  "myapp",
				},
				Items: []*itemConfig{
					{Name: "inst-abc123"},
					{Name: "app-xyz789"},
				},
				Labels: map[string]string{
					"install": "inst-abc123",
					"app":     "app-xyz789",
				},
				Data: []byte("acme-corp"),
			},
			shouldError: false,
		},
		"multiple variables in one field": {
			input: &simpleStruct{
				ID: "{{.nuon.install.id}}-{{.nuon.app.id}}-{{.nuon.org.id}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "inst-abc123-app-xyz789-org-123456",
			},
			shouldError: false,
		},
		"sprig function usage - upper": {
			input: &simpleStruct{
				ID: "{{.nuon.install.id | upper}}",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "INST-ABC123",
			},
			shouldError: false,
		},
		"real world kubernetes manifest config": {
			input: &realWorldConfig{
				Manifest:  "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: {{.nuon.install.namespace}}",
				Namespace: "{{.nuon.install.namespace}}",
				Kustomize: &kustomizeConfig{
					Path: "overlays/{{.nuon.install.name}}",
				},
			},
			data: getStandardData(),
			expected: &realWorldConfig{
				Manifest:  "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: prod-ns",
				Namespace: "prod-ns",
				Kustomize: &kustomizeConfig{
					Path: "overlays/production",
				},
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_Errors() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		shouldError bool
		errorMsg    string
	}{
		"missing required nuon data": {
			input: &simpleStruct{
				ID: "{{.nuon.missing.value}}",
			},
			data:        getMinimalData(),
			shouldError: true,
		},
		"invalid template syntax - unclosed": {
			input: &simpleStruct{
				ID: "{{.nuon.install.id}",
			},
			data:        getStandardData(),
			shouldError: true,
		},
		"nested struct with missing data": {
			input: &nestedStruct{
				Parent: "{{.nuon.install.id}}",
				Child: &childStruct{
					Value: "{{.nuon.missing.field}}",
				},
			},
			data:        getMinimalData(),
			shouldError: true,
		},
		"slice element with missing data": {
			input: &sliceStruct{
				Items: []string{
					"{{.nuon.install.id}}",
					"{{.nuon.missing.value}}",
				},
			},
			data:        getMinimalData(),
			shouldError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStruct_EdgeCases() {
	t := s.T()
	tests := map[string]struct {
		input       any
		data        map[string]any
		expected    any
		shouldError bool
	}{
		"unexported fields ignored": {
			input: &unexportedFieldStruct{
				Public:  "{{.nuon.install.id}}",
				private: "{{.nuon.install.name}}",
			},
			data: getStandardData(),
			expected: &unexportedFieldStruct{
				Public:  "inst-abc123",
				private: "{{.nuon.install.name}}",
			},
			shouldError: false,
		},
		"nil input pointer": {
			input: &nestedStruct{
				Parent: "{{.nuon.install.id}}",
				Child:  nil,
			},
			data: getStandardData(),
			expected: &nestedStruct{
				Parent: "inst-abc123",
				Child:  nil,
			},
			shouldError: false,
		},
		"empty data with non-nuon template": {
			input: &simpleStruct{
				ID: "{{.other.value}}",
			},
			data: map[string]any{},
			expected: &simpleStruct{
				ID: "{{.other.value}}",
			},
			shouldError: false,
		},
		"unicode in template values": {
			input: &simpleStruct{
				ID: "prefix-{{.nuon.install.id}}-后缀",
			},
			data: getStandardData(),
			expected: &simpleStruct{
				ID: "prefix-inst-abc123-后缀",
			},
			shouldError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := RenderStruct(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderStrField() {
	t := s.T()
	tests := map[string]struct {
		input       string
		data        map[string]any
		expected    string
		shouldError bool
	}{
		"simple template": {
			input:    "{{.nuon.install.id}}",
			data:     getStandardData(),
			expected: "inst-abc123",
		},
		"template with prefix": {
			input:    "prefix-{{.nuon.install.id}}",
			data:     getStandardData(),
			expected: "prefix-inst-abc123",
		},
		"no template": {
			input:    "static-value",
			data:     getStandardData(),
			expected: "static-value",
		},
		"missing data": {
			input:       "{{.nuon.missing.field}}",
			data:        getMinimalData(),
			shouldError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := renderStrField(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func (s *RenderStructTestSuite) TestRenderByteField() {
	t := s.T()
	tests := map[string]struct {
		input       []byte
		data        map[string]any
		expected    []byte
		shouldError bool
	}{
		"simple template": {
			input:    []byte("{{.nuon.install.id}}"),
			data:     getStandardData(),
			expected: []byte("inst-abc123"),
		},
		"template with prefix": {
			input:    []byte("prefix-{{.nuon.install.id}}"),
			data:     getStandardData(),
			expected: []byte("prefix-inst-abc123"),
		},
		"no template": {
			input:    []byte("static-value"),
			data:     getStandardData(),
			expected: []byte("static-value"),
		},
		"missing data": {
			input:       []byte("{{.nuon.missing.field}}"),
			data:        getMinimalData(),
			shouldError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := renderByteField(tc.input, tc.data)

			if tc.shouldError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}
