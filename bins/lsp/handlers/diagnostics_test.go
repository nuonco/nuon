package handlers

import (
	"math"
	"strings"
	"testing"

	"github.com/invopop/jsonschema"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
)

// TestDiagnostics_TerraformVersionTypeMismatch tests that type mismatches are caught
// for both valid floats (1.2) and invalid floats (1.2.3) in terraform_version fields.
func TestDiagnostics_TerraformVersionTypeMismatch(t *testing.T) {
	// Create a minimal schema that expects terraform_version to be a string
	props := jsonschema.NewProperties()
	props.Set("terraform_version", &jsonschema.Schema{
		Type: "string",
	})

	schema := &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Required:    []string{},
		Definitions: map[string]*jsonschema.Schema{},
	}

	testCases := []struct {
		name            string
		tomlContent     string
		shouldHaveError bool
		errorMessage    string
	}{
		{
			name:            "Valid TOML float 1.2 should error",
			tomlContent:     "terraform_version = 1.2",
			shouldHaveError: true,
			errorMessage:    "Type mismatch for 'terraform_version'",
		},
		{
			name:            "Invalid TOML float 1.2.3 should error",
			tomlContent:     "terraform_version = 1.2.3",
			shouldHaveError: true,
			errorMessage:    "Type mismatch for 'terraform_version'",
		},
		{
			name:            "String value should not error",
			tomlContent:     `terraform_version = "1.2.3"`,
			shouldHaveError: false,
		},
		{
			name:            "Integer should error",
			tomlContent:     "terraform_version = 42",
			shouldHaveError: true,
			errorMessage:    "Type mismatch for 'terraform_version'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse TOML
			doc := tomlparser.ParseToml(tc.tomlContent)

			// Extract values (this is what PublishDiagnostics does)
			if strictDoc, err := tomlparser.ParseStrict(tc.tomlContent); err == nil {
				doc.Values = strictDoc.Values
			} else {
				// Use our new extractRawValues fallback
				doc.Values = extractRawValues(tc.tomlContent, doc)
			}

			// Generate diagnostics
			diags := DiagnoseDocument("file:///test.toml", doc, schema)

			// Check results
			hasTypeMismatch := false
			for _, diag := range diags {
				if tc.errorMessage != "" && strings.Contains(diag.Message, tc.errorMessage) {
					hasTypeMismatch = true
					break
				}
			}

			if tc.shouldHaveError && !hasTypeMismatch {
				t.Errorf("Expected type mismatch error for '%s', but got none. Diagnostics: %v", tc.tomlContent, diags)
			}

			if !tc.shouldHaveError && hasTypeMismatch {
				t.Errorf("Expected no type mismatch error for '%s', but got one: %v", tc.tomlContent, diags)
			}
		})
	}
}

// TestExtractRawValues_ParsesValuesFromInvalidTOML tests that extractRawValues
// correctly extracts values even when TOML is syntactically invalid
func TestExtractRawValues_ParsesValuesFromInvalidTOML(t *testing.T) {
	testCases := []struct {
		name          string
		tomlContent   string
		expectedKey   string
		expectedValue any
	}{
		{
			name:          "Should parse valid float",
			tomlContent:   "terraform_version = 1.2",
			expectedKey:   "terraform_version",
			expectedValue: 1.2,
		},
		{
			name:          "Should parse invalid float as numeric (0.0) for type checking",
			tomlContent:   "terraform_version = 1.2.3",
			expectedKey:   "terraform_version",
			expectedValue: 0.0,
		},
		{
			name:          "Should parse integer",
			tomlContent:   "port = 8080",
			expectedKey:   "port",
			expectedValue: int64(8080),
		},
		{
			name:          "Should parse string",
			tomlContent:   `url = https://example.com`,
			expectedKey:   "url",
			expectedValue: "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := tomlparser.ParseToml(tc.tomlContent)
			values := extractRawValues(tc.tomlContent, doc)

			val, exists := values[tc.expectedKey]
			if !exists {
				t.Fatalf("Expected key '%s' to be extracted, but it wasn't. Got: %v", tc.expectedKey, values)
			}

			if f, ok := tc.expectedValue.(float64); ok {
				if fVal, ok := val.(float64); ok {
					const eps = 1e-9
					if math.Abs(fVal-f) > eps {
						t.Errorf("Expected value %v, got %v", tc.expectedValue, val)
					}
				} else {
					t.Errorf("Expected float64, got %T with value %v", val, val)
				}
			} else if tc.expectedValue != val {
				t.Errorf("Expected value %v, got %v", tc.expectedValue, val)
			}
		})
	}
}

// TestDiagnostics_MissingRequiredFields tests that missing required fields are detected
func TestDiagnostics_MissingRequiredFields(t *testing.T) {
	props := jsonschema.NewProperties()
	props.Set("name", &jsonschema.Schema{Type: "string"})
	props.Set("type", &jsonschema.Schema{Type: "string"})

	schema := &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Required:    []string{"name", "type"},
		Definitions: map[string]*jsonschema.Schema{},
	}

	tomlContent := `type = "test"`
	doc := tomlparser.ParseToml(tomlContent)

	diags := DiagnoseDocument("file:///test.toml", doc, schema)

	hasMissingName := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "name") && strings.Contains(diag.Message, "required") {
			hasMissingName = true
			break
		}
	}

	if !hasMissingName {
		t.Errorf("Expected diagnostic for missing required field 'name', got: %v", diags)
	}
}

// TestDiagnostics_UnknownKeys tests that unknown keys are detected
func TestDiagnostics_UnknownKeys(t *testing.T) {
	props := jsonschema.NewProperties()
	props.Set("name", &jsonschema.Schema{Type: "string"})

	schema := &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Definitions: map[string]*jsonschema.Schema{},
	}

	tomlContent := `name = "test"
unknown_key = "value"`
	doc := tomlparser.ParseToml(tomlContent)

	diags := DiagnoseDocument("file:///test.toml", doc, schema)

	hasUnknownKey := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "unknown_key") && strings.Contains(diag.Message, "Unknown") {
			hasUnknownKey = true
			break
		}
	}

	if !hasUnknownKey {
		t.Errorf("Expected diagnostic for unknown key, got: %v", diags)
	}
}
