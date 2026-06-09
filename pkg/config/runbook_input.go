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
	// NewSchemaBuilder(schema).
	// 	Field("name").Short("input name").
	// 	Long("Used to reference the input within runbook step fields via templating (e.g., {{.runbook_inputs.input_name}})").
	// 	Example("image_tag").
	// 	Example("target_env").
	// 	Field("display_name").Short("display name of the input").Required().
	// 	Long("Human-readable name shown in the run-runbook dialog").
	// 	Example("Image Tag").
	// 	Field("description").Short("input description").Required().
	// 	Long("Detailed explanation of what this input is for, shown when running the runbook").
	// 	Example("The container image tag to deploy").
	// 	Field("default").Short("default value for the input").
	// 	Long("Default value used if the operator does not provide one when running the runbook").
	// 	Example("latest").
	// 	Field("required").Short("whether input is required").
	// 	Long("If true, a value must be provided before the runbook can run").
	// 	Field("sensitive").Short("whether input is sensitive").
	// 	Long("If true, the value is masked/redacted when the run is read back. Use for tokens and secrets").
	// 	Field("type").Short("input type").
	// 	Long("Data type for the input. Supported types: string, number, list, json, bool").
	// 	Example("string").
	// 	Example("number")
}
