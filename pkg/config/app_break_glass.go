package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type BreakGlass struct {
	Roles []*AppAWSIAMRole `json:"Roles" mapstructure:"role,omitempty" toml:"role,omitempty"`
}

func (a BreakGlass) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "role", "Roles to be used for breaking glass.")
}

func (a *BreakGlass) parse(context.Context) error {
	return nil
}

func (a *BreakGlass) Validate() error {
	if a == nil {
		return nil
	}

	for _, role := range a.Roles {
		if role == nil {
			continue
		}
		if err := ValidateRoleGrants(role.Name, role); err != nil {
			return err
		}
	}

	return nil
}
