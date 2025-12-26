package config

import (
	"github.com/invopop/jsonschema"
)

type AppAWSIAMRole struct {
	Type string `mapstructure:"type" toml:"type"`

	Name        string            `mapstructure:"name" toml:"name" jsonschema:"required" features:"template"`
	Description string            `mapstructure:"description" toml:"description" jsonschema:"required" features:"template"`
	DisplayName string            `mapstructure:"display_name,omitempty" toml:"display_name,omitempty" features:"template"`
	Policies    []AppAWSIAMPolicy `mapstructure:"policies" toml:"policies" jsonschema:"required"`

	PermissionsBoundary string `mapstructure:"permissions_boundary,omitempty" toml:"permissions_boundary,omitempty" features:"template,get"`
}

func (a AppAWSIAMRole) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("role type in permission directory").
		Long("Used when defining permissions in a directory. Indicates when the role is active (provision, maintenance, or deprovision). Supports templating").
		Example("provision").
		Example("maintenance").
		Example("deprovision").
		Field("name").Short("name of the IAM role").Required().
		Long("Name used for the role in AWS. Supports Go templating using standard template variables (e.g., {{.nuon.install.id}})").
		Example("app-{{.nuon.install.id}}-role").
		Example("admin-role").
		Field("description").Short("description of the role").Required().
		Long("Human-readable description that explains the role's purpose. Rendered in the installer to customers. Supports templating").
		Example("Provides S3 bucket access for the application").
		Example("Database migration role with elevated permissions").
		Field("display_name").Short("display name of the role").
		Long("Human-readable display name shown in the installer UI. Supports templating").
		Example("Application S3 Access").
		Example("Database Admin").
		Field("policies").Short("policy definitions for the role").Required().
		Long("List of IAM policies to attach to the role. Each policy defines specific AWS permissions").
		Field("permissions_boundary").Short("permissions boundary policy").
		Long("Optional ARN of a permissions boundary policy. Limits the maximum permissions the role can have. Supports templating and external file sources: HTTP(S) URLs (https://example.com/boundary.json), git repositories (git::https://github.com/org/repo//boundary.json), file paths (file:///path/to/boundary.json), and relative paths (./boundary.json)")
}
