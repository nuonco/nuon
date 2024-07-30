package config

import "github.com/invopop/jsonschema"

func addDescription(schema *jsonschema.Schema, name, description string) {
	field, ok := schema.Properties.Get(name)
	if ok {
		field.Description = description
	}
}
