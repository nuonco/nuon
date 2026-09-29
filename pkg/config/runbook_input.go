package config

import (
	"github.com/invopop/jsonschema"
)

type RunbookInput struct {
	Name        string `mapstructure:"name" toml:"name"`
	DisplayName string `mapstructure:"display_name" toml:"display_name" jsonschema:"required"`
	Description string `mapstructure:"description" toml:"description" jsonschema:"required"`
	Default     any    `mapstructure:"default,omitempty" toml:"default,omitempty"`
	Required    bool   `mapstructure:"required,omitempty" toml:"required,omitempty"`
	Sensitive   bool   `mapstructure:"sensitive" toml:"sensitive"`
	Type        string `mapstructure:"type" toml:"type"`
}

func (r RunbookInput) JSONSchemaExtend(schema *jsonschema.Schema) {
}
