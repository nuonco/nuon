package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type StackConfig struct {
	Type        string `mapstructure:"type"`
	Name        string `mapstructure:"name" jsonschema:"required" features:"template"`
	Description string `mapstructure:"description" jsonschema:"required" features:"template"`

	VPCNestedTemplateURL    string `mapstructure:"vpc_nested_template_url"`
	RunnerNestedTemplateURL string `mapstructure:"runner_nested_template_url"`
}

func (a StackConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "type", "Stack type (must be aws-cloudformation).")
	addDescription(schema, "name", "Name of the stack when installed in the customer account. Supports templating.")
	addDescription(schema, "description", "Description of the stack.")
	addDescription(schema, "vpc_nested_template_url", "Nested template used for VPC.")
	addDescription(schema, "runner_nested_template_url", "Nested template used for runner.")
}

func (a *StackConfig) parse(context.Context) error {
	return nil
}
