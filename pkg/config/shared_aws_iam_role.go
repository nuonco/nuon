package config

import (
	"github.com/invopop/jsonschema"
)

type AppAWSIAMRole struct {
	Name        string            `mapstructure:"name" jsonschema:"required" features:"template"`
	Description string            `mapstructure:"description" jsonschema:"required" features:"template"`
	DisplayName string            `mapstructure:"display_name,omitempty" features:"template"`
	Policies    []AppAWSIAMPolicy `mapstructure:"policies" jsonschema:"required"`

	PermissionsBoundary string `mapstructure:"permissions_boundary,omitempty" features:"template,get"`
}

func (a AppAWSIAMRole) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "Name used for the role, this can be templated using standard template variables.")
	addDescription(schema, "description", "Human readable description which is rendered in the installer.")
	addDescription(schema, "display_name", "Human readable name which is rendered in the installer.")
	addDescription(schema, "policies", "Policy definitions that should be created on the role.")
}
