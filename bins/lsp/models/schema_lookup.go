package models

import (
	"bufio"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/schema"
)

func DetectSchemaType(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" {
				return comment
			}
		}

		if !strings.HasPrefix(line, "#") {
			break
		}
	}
	return ""
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
