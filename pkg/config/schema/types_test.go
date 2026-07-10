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

// TestPermissionVsPermissionsSchemas guards the singular/collection split: the
// permissions/ directory form is a single AppAWSIAMRole per file ("permission"),
// while permissions.toml is the PermissionsConfig collection ("permissions").
// They must be distinct schemas with distinct $ids.
func TestPermissionVsPermissionsSchemas(t *testing.T) {
	single, err := LookupSchemaType("permission")
	if err != nil || single == nil {
		t.Fatalf("permission schema unavailable: %v", err)
	}
	collection, err := LookupSchemaType("permissions")
	if err != nil || collection == nil {
		t.Fatalf("permissions schema unavailable: %v", err)
	}
	if single.ID == "" {
		t.Fatal("permission schema must declare a root $id")
	}
	if single.ID == collection.ID {
		t.Fatalf("permission and permissions must not share $id %q", single.ID)
	}
}

func TestLookupSchemaTypeNormalizesUnderscores(t *testing.T) {
	tests := []struct {
		typ   string
		found bool
	}{
		{"container-image", true},
		{"container_image", true},
		{"docker_build", true},
		{"job", true},
		{"runner", true},
		{"unknown-type", false},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			schm, err := LookupSchemaType(tt.typ)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.typ, err)
			}
			if tt.found && schm == nil {
				t.Fatalf("expected schema for %s, got nil", tt.typ)
			}
			if !tt.found && schm != nil {
				t.Fatalf("expected no schema for %s", tt.typ)
			}
			if got := IsValidSchemaType(tt.typ); got != tt.found {
				t.Fatalf("IsValidSchemaType(%s) = %v, want %v", tt.typ, got, tt.found)
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
