package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"

	"github.com/nuonco/nuon/pkg/config"
)

// mockResultError implements gojsonschema.ResultError for testing
type mockResultError struct {
	gojsonschema.ResultErrorFields
}

func newMockError(field, description string) *mockResultError {
	err := &mockResultError{}
	err.SetType("invalid_type")
	err.SetDescription(description)
	// Use SetContext to set the field path
	ctx := gojsonschema.NewJsonContext(field, nil)
	err.SetContext(ctx)
	return err
}

func TestFormatValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		errors     []gojsonschema.ResultError
		components []*config.Component
		expected   []string
	}{
		{
			name:       "empty errors",
			errors:     []gojsonschema.ResultError{},
			components: nil,
			expected:   []string{},
		},
		{
			name: "non-component error unchanged",
			errors: []gojsonschema.ResultError{
				newMockError("metadata.name", "is required"),
			},
			components: nil,
			expected:   []string{"metadata.name: is required"},
		},
		{
			name: "component error with source file",
			errors: []gojsonschema.ResultError{
				newMockError("components.0", "Invalid type. Expected: object, given: null"),
			},
			components: []*config.Component{
				{Name: "my-component", SourceFile: "components/my-component.toml"},
			},
			expected: []string{"components/my-component.toml: Invalid type. Expected: object, given: null"},
		},
		{
			name: "component nested error with source file",
			errors: []gojsonschema.ResultError{
				newMockError("components.1.helm_chart.chart_name", "is required"),
			},
			components: []*config.Component{
				{Name: "first", SourceFile: "components/first.toml"},
				{Name: "second", SourceFile: "components/second.toml"},
			},
			expected: []string{"components/second.toml.helm_chart.chart_name: is required"},
		},
		{
			name: "component error without source file falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("components.0", "Invalid type"),
			},
			components: []*config.Component{
				{Name: "my-component", SourceFile: ""},
			},
			expected: []string{"components.0: Invalid type"},
		},
		{
			name: "component error with nil component falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("components.0", "Invalid type"),
			},
			components: []*config.Component{nil},
			expected:   []string{"components.0: Invalid type"},
		},
		{
			name: "component error with out-of-range index falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("components.5", "Invalid type"),
			},
			components: []*config.Component{
				{Name: "only-one", SourceFile: "components/only-one.toml"},
			},
			expected: []string{"components.5: Invalid type"},
		},
		{
			name: "multiple errors mixed",
			errors: []gojsonschema.ResultError{
				newMockError("components.0", "Invalid type"),
				newMockError("metadata.version", "is required"),
				newMockError("components.1.name", "is required"),
			},
			components: []*config.Component{
				{Name: "first", SourceFile: "components/first.toml"},
				{Name: "second", SourceFile: "components/second.toml"},
			},
			expected: []string{
				"components/first.toml: Invalid type",
				"metadata.version: is required",
				"components/second.toml.name: is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AppConfig{
				Components: tt.components,
			}
			result := formatValidationErrors(tt.errors, cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidationErrors_Policies(t *testing.T) {
	tests := []struct {
		name     string
		errors   []gojsonschema.ResultError
		policies *config.PoliciesConfig
		expected []string
	}{
		{
			name: "policy error with source file and line number",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.0.type", "must be one of the following: kubernetes_cluster, terraform_module"),
			},
			policies: &config.PoliciesConfig{
				Policies: []config.AppPolicy{
					{Type: "invalid", SourceFile: "policies/my-policy.toml", SourceLine: 15},
				},
			},
			expected: []string{"policies/my-policy.toml:L15.type: must be one of the following: kubernetes_cluster, terraform_module"},
		},
		{
			name: "policy error with source file but no line number falls back to policy number",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.0.type", "must be one of the following: kubernetes_cluster, terraform_module"),
			},
			policies: &config.PoliciesConfig{
				Policies: []config.AppPolicy{
					{Type: "invalid", SourceFile: "policies/my-policy.toml"},
				},
			},
			expected: []string{"policies/my-policy.toml (policy 1).type: must be one of the following: kubernetes_cluster, terraform_module"},
		},
		{
			name: "policy error without source file falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.0.type", "is required"),
			},
			policies: &config.PoliciesConfig{
				Policies: []config.AppPolicy{
					{Type: "kubernetes_cluster", SourceFile: ""},
				},
			},
			expected: []string{"policies.policy.0.type: is required"},
		},
		{
			name: "policy nested error with source file and line number",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.2.engine", "must be one of the following: kyverno, opa"),
			},
			policies: &config.PoliciesConfig{
				Policies: []config.AppPolicy{
					{Type: "kubernetes_cluster", SourceFile: "policies/first.toml", SourceLine: 1},
					{Type: "terraform_module", SourceFile: "policies/second.toml", SourceLine: 10},
					{Type: "helm_chart", SourceFile: "policies/third.toml", SourceLine: 25},
				},
			},
			expected: []string{"policies/third.toml:L25.engine: must be one of the following: kyverno, opa"},
		},
		{
			name: "policy error with out-of-range index falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.10.type", "Invalid type"),
			},
			policies: &config.PoliciesConfig{
				Policies: []config.AppPolicy{
					{Type: "kubernetes_cluster", SourceFile: "policies/first.toml"},
				},
			},
			expected: []string{"policies.policy.10.type: Invalid type"},
		},
		{
			name: "policy error with nil policies config falls back to index",
			errors: []gojsonschema.ResultError{
				newMockError("policies.policy.0.type", "Invalid type"),
			},
			policies: nil,
			expected: []string{"policies.policy.0.type: Invalid type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.AppConfig{
				Policies: tt.policies,
			}
			result := formatValidationErrors(tt.errors, cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
