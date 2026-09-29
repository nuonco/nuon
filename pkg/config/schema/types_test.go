package schema

import (
	"fmt"
	"testing"

	"github.com/invopop/jsonschema"
)

func TestAllSchemasHaveJSONSchemaExtend(t *testing.T) {
	tests := make([]struct {
		name string
		fn   func() (*string, error)
	}, 0, len(SchemaMapping))

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

func TestPermissionPolicySchema(t *testing.T) {
	schm, err := LookupSchemaType("permission-policy")
	if err != nil || schm == nil {
		t.Fatalf("permission-policy schema unavailable: %v", err)
	}
	if schm.ID == "" {
		t.Fatal("permission-policy schema must declare a root $id")
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

func TestValidateJSONSchemaExtendSucceedsWithValidStruct(t *testing.T) {
	err := ValidateJSONSchemaExtend(TestValidatorStruct{})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

type TestValidatorStruct struct {
	Field string
}

func (t TestValidatorStruct) JSONSchemaExtend(schema *jsonschema.Schema) {
}
