package schema

import (
	"fmt"
	"testing"

	"github.com/invopop/jsonschema"
)

// TestAllSchemasHaveJSONSchemaExtend ensures all schema functions validate successfully.
// This test protects against regressions where new struct fields are added without
// implementing JSONSchemaExtend on nested types.
func TestAllSchemasHaveJSONSchemaExtend(t *testing.T) {
	tests := make([]struct {
		name string
		fn   func() (*string, error)
	}, 0, len(SchemaMapping))

	// Convert SchemaMapping to test cases
	for schemaType, schemaFn := range SchemaMapping {
		schemaType := schemaType
		schemaFn := schemaFn

		tests = append(tests, struct {
			name string
			fn   func() (*string, error)
		}{
			name: schemaType,
			fn: func() (*string, error) {
				_, err := schemaFn()
				if err != nil {
					return nil, err
				}
				return nil, nil
			},
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fn()
			if err != nil {
				t.Fatalf("schema %s failed validation: %v", tt.name, err)
			}
		})
	}
}

// TestValidateJSONSchemaExtendOnMissingImplementation verifies that the validator
// correctly detects when a struct doesn't implement JSONSchemaExtend.
func TestValidateJSONSchemaExtendDetectsMissing(t *testing.T) {
	type MissingJSONSchemaExtend struct {
		Field string
	}

	err := ValidateJSONSchemaExtend(MissingJSONSchemaExtend{})
	if err == nil {
		t.Fatalf("expected validation error for struct without JSONSchemaExtend, got nil")
	}

	if err.Error() != fmt.Sprintf("struct %s does not implement JSONSchemaExtend(*jsonschema.Schema)", "MissingJSONSchemaExtend") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// TestValidateJSONSchemaExtendSucceedsWithValidStruct verifies that the validator
// passes for properly implemented structs.
func TestValidateJSONSchemaExtendSucceedsWithValidStruct(t *testing.T) {
	// Use an existing config struct that has JSONSchemaExtend implemented
	err := ValidateJSONSchemaExtend(TestValidatorStruct{})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestValidatorStruct is a test struct with JSONSchemaExtend for validator testing
type TestValidatorStruct struct {
	Field string
}

func (t TestValidatorStruct) JSONSchemaExtend(schema *jsonschema.Schema) {
	// No-op implementation for testing
}
