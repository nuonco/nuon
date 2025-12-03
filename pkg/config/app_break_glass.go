package config

import (
	"github.com/invopop/jsonschema"
)

type BreakGlass struct {
	Roles []*AppAWSIAMRole `mapstructure:"role,omitempty" toml:"role,omitempty"`
}

func (a BreakGlass) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "role", "Roles to be used for breaking glass.")
}
