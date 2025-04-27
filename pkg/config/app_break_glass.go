package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type BreakGlass struct {
	Roles []*AppAWSIAMRole `mapstructure:"role,omitempty"`
}

func (a BreakGlass) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "role", "Roles to be used for breaking glass.")
}

func (a *BreakGlass) parse(context.Context) error {
	return nil
}
