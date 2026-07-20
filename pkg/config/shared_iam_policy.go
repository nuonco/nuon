package config

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
)

type AppAWSIAMPolicy struct {
	ManagedPolicyName string `mapstructure:"managed_policy_name,omitempty" toml:"managed_policy_name,omitempty"`
	// Name is optional: a managed_policy_name attachment identifies itself, so a
	// bare AWS managed policy needs no separate name. The runtime does not
	// require it (see parse below).
	Name string `mapstructure:"name,omitempty" toml:"name,omitempty"`

	Contents string `mapstructure:"contents" toml:"contents" features:"template,get"`

	GCPPermissions    []string `mapstructure:"gcp_permissions,omitempty" toml:"gcp_permissions,omitempty"`
	GCPPredefinedRole string   `mapstructure:"gcp_predefined_role,omitempty" toml:"gcp_predefined_role,omitempty"`

	AzureActions      []string `mapstructure:"azure_actions,omitempty" toml:"azure_actions,omitempty"`
	AzureBuiltInRoles []string `mapstructure:"azure_built_in_roles,omitempty" toml:"azure_built_in_roles,omitempty"`
}

func (a AppAWSIAMPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("policy name").
		Long("Name for the policy. Used across all cloud platforms when creating the permission grant. Optional for a bare AWS managed_policy_name attachment, which needs no separate name. Supports Nuon templating").
		Example("app-{{.nuon.install.id}}-policy").
		Example("s3-access-policy").
		Field("managed_policy_name").Short("[AWS] managed policy name").OneOfRequired("aws_policy").
		Long("[AWS only] Name or ARN of an AWS managed policy to attach to the IAM role. Mutually exclusive with contents").
		Example("AmazonS3FullAccess").
		Example("ReadOnlyAccess").
		Field("contents").Short("[AWS] inline policy document").OneOfRequired("aws_policy").
		Long("[AWS only] JSON policy document defining inline IAM permissions. Mutually exclusive with managed_policy_name. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/policy.json), git repositories (git::https://github.com/org/repo//policy.json), file paths (file:///path/to/policy.json), and relative paths (./policy.json)").
		Example("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"s3:*\",\"Resource\":\"*\"}]}").
		Field("gcp_permissions").Short("[GCP] individual permissions").OneOfRequired("gcp_policy").
		Long("[GCP only] List of individual GCP IAM permission strings to include in a custom role bound to the service account. Use this for fine-grained permission control. Mutually exclusive with gcp_predefined_role").
		Example("compute.instances.get").
		Example("storage.objects.list").
		Field("gcp_predefined_role").Short("[GCP] predefined role").OneOfRequired("gcp_policy").
		Long("[GCP only] Name of a GCP predefined role to bind to the service account. This is the GCP equivalent of AWS managed policies — a Google-managed bundle of permissions. Mutually exclusive with gcp_permissions").
		Example("roles/editor").
		Example("roles/owner").
		Field("azure_actions").Short("[Azure] individual RBAC actions").OneOfRequired("azure_policy").
		Long("[Azure only] List of Azure RBAC action strings to include in a custom role definition bound to the operation's managed identity. Use this for fine-grained permission control. Mutually exclusive with azure_built_in_roles").
		Example("Microsoft.Compute/*").
		Example("Microsoft.Resources/subscriptions/resourceGroups/*").
		Field("azure_built_in_roles").Short("[Azure] built-in roles").OneOfRequired("azure_policy").
		Long("[Azure only] Names of Azure built-in roles to assign to the operation's managed identity (e.g. Contributor, Reader). This is the Azure equivalent of AWS managed policies. Mutually exclusive with azure_actions").
		Example("Contributor").
		Example("Reader")
}

func (a *AppAWSIAMPolicy) parse(ctx context.Context) error {
	if a.Contents != "" && a.ManagedPolicyName != "" {
		return fmt.Errorf("policy %q: contents and managed_policy_name are mutually exclusive; specify one or the other", a.Name)
	}

	if len(a.GCPPermissions) > 0 && a.GCPPredefinedRole != "" {
		return fmt.Errorf("policy %q: gcp_permissions and gcp_predefined_role are mutually exclusive; use gcp_permissions for fine-grained custom permissions or gcp_predefined_role for a Google-managed role, not both", a.Name)
	}

	if len(a.AzureActions) > 0 && len(a.AzureBuiltInRoles) > 0 {
		return fmt.Errorf("policy %q: azure_actions and azure_built_in_roles are mutually exclusive; use azure_actions for a fine-grained custom role or azure_built_in_roles for Azure-managed roles, not both", a.Name)
	}

	return nil
}
