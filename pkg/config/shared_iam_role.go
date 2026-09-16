package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type AppAWSIAMRole struct {
	Type                string            `json:"Type" mapstructure:"type" toml:"type"`
	CloudPlatform       string            `json:"CloudPlatform" mapstructure:"cloud_platform,omitempty" toml:"cloud_platform,omitempty"`
	Name                string            `json:"Name" mapstructure:"name" toml:"name" jsonschema:"required" features:"template"`
	Description         string            `json:"Description" mapstructure:"description" toml:"description" jsonschema:"required" features:"template"`
	DisplayName         string            `json:"DisplayName" mapstructure:"display_name,omitempty" toml:"display_name,omitempty" features:"template"`
	Policies            []AppAWSIAMPolicy `json:"Policies" mapstructure:"policies" toml:"policies" jsonschema:"required"`
	PermissionsBoundary string            `json:"PermissionsBoundary" mapstructure:"permissions_boundary,omitempty" toml:"permissions_boundary,omitempty" features:"template,get"`
	EnabledInStack      *bool             `json:"EnabledInStack" mapstructure:"enabled_in_stack,omitempty" toml:"enabled_in_stack,omitempty"`
	NamedPolicies       []NamedPolicyRef  `json:"NamedPolicies" mapstructure:"named_policies,omitempty" toml:"named_policies,omitempty"`
}

func (a AppAWSIAMRole) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("role type in permission directory").
		Long("Used when defining permissions in a directory. Indicates when the role is active (provision, maintenance, or deprovision). Supports templating").
		Example("provision").
		Example("maintenance").
		Example("deprovision").
		Field("cloud_platform").Short("target cloud platform").
		Long("Cloud platform this role targets. Determines which downstream renderer processes the role (e.g., AWS CloudFormation vs GCP IAM). Defaults to aws if omitted").
		Enum("aws", "azure", "gcp").
		Example("aws").
		Example("gcp").
		Field("name").Short("name of the role").Required().
		Long("Name used for the role in the target cloud platform. Supports Go templating using standard template variables (e.g., {{.nuon.install.id}})").
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
		Field("policies").Short("policy definitions for the role").
		Long("List of policies to attach to the role. Each policy defines cloud-specific permissions (AWS IAM policies, GCP IAM permissions, or GCP predefined roles). May be omitted when the role attaches at least one named_policies entry, which carries the grants instead").
		Field("permissions_boundary").Short("[AWS] permissions boundary policy").
		Long("[AWS only] Optional ARN of a permissions boundary policy. Limits the maximum permissions the role can have. Supports templating and external file sources: HTTP(S) URLs (https://example.com/boundary.json), git repositories (git::https://github.com/org/repo//boundary.json), file paths (file:///path/to/boundary.json), and relative paths (./boundary.json)").
		Example("./provision_boundary.json").
		Example("./maintenance_boundary.json").
		Field("enabled_in_stack").Short("whether the role is enabled by default in the install stack").
		Long("Controls the default value of the Enable parameter for this role in the install stack (CloudFormation parameter or Terraform variable, depending on which format the customer applies). When true, the role is created by default. When false, the role is not created unless the installer explicitly enables it. If omitted, the platform default is used (true for standard roles, false for break-glass roles)").
		Field("named_policies").Short("named IAM policies to attach").
		Long("Named IAM policies to attach to this role, referenced by name. Use [[named_policies]] like [[policies]]. The name must match a named policy defined under permissions. Those policies are created even when this role is disabled")
}

// ValidateRoleGrants rejects a role that grants nothing. policies is empty only
// when named_policies carries the grants instead, so the two are checked
// together rather than with a jsonschema required tag on either one.
func ValidateRoleGrants(label string, role *AppAWSIAMRole) error {
	if role == nil {
		return nil
	}

	if len(role.Policies) > 0 || len(role.NamedPolicies) > 0 {
		return nil
	}

	return fmt.Errorf("role %q has no permissions: set at least one [[policies]] or [[named_policies]] entry", label)
}
