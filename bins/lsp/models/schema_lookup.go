package models

import (
	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/diagnostics"
	"github.com/nuonco/nuon/pkg/config/schema"
)

func DetectSchemaType(text string) string {
	return diagnostics.DetectSchemaType(text)
}

func DetectSchemaTypeForDocument(text, uri string) string {
	return diagnostics.DetectSchemaTypeForDocument(text, uri)
}

func LookupSchema(schemaType string) (*jsonschema.Schema, error) {
	return schema.LookupSchemaType(schemaType)
}

// GetValidSchemaTypes returns all valid schema type names
func GetValidSchemaTypes() []string {
	return schema.GetSchemaTypes()
}

// IsValidSchemaType checks if a schema type is valid
func IsValidSchemaType(schemaType string) bool {
	return schema.IsValidSchemaType(schemaType)
}
