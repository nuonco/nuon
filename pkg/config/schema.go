package config

import "github.com/invopop/jsonschema"

func addDescription(schema *jsonschema.Schema, name, description string) {
	field, ok := schema.Properties.Get(name)
	if ok {
		field.Description = description
	}
}

// markDeprecated is kept for potential future use.
// nolint:unused
func markDeprecated(schema *jsonschema.Schema, name string, deprecationMessage string) {
	field, ok := schema.Properties.Get(name)
	if ok {
		field.Deprecated = true
		if deprecationMessage != "" {
			if field.Description != "" {
				field.Description = field.Description + " [DEPRECATED: " + deprecationMessage + "]"
			} else {
				field.Description = "[DEPRECATED: " + deprecationMessage + "]"
			}
		}
	}
}
