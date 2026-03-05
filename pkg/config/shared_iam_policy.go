package config

import (
	"context"
	"fmt"

	"github.com/invopop/jsonschema"
)

type AppAWSIAMPolicy struct {
	ManagedPolicyName string `mapstructure:"managed_policy_name,omitempty" toml:"managed_policy_name,omitempty"`
	Name              string `mapstructure:"name" toml:"name" jsonschema:"required"`

	Contents string `mapstructure:"contents" toml:"contents" features:"template,get"`

	GCPPermissions    []string `mapstructure:"gcp_permissions,omitempty" toml:"gcp_permissions,omitempty"`
	GCPPredefinedRole string   `mapstructure:"gcp_predefined_role,omitempty" toml:"gcp_predefined_role,omitempty"`
}

func (a AppAWSIAMPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("policy name").Required().
		Long("Name for the policy. Used across all cloud platforms when creating the permission grant. Supports Nuon templating").
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
		Example("roles/owner")
}

func (a *AppAWSIAMPolicy) parse(ctx context.Context) error {
	if a.Contents != "" && a.ManagedPolicyName != "" {
		return fmt.Errorf("policy %q: contents and managed_policy_name are mutually exclusive; specify one or the other", a.Name)
	}

	if len(a.GCPPermissions) > 0 && a.GCPPredefinedRole != "" {
		return fmt.Errorf("policy %q: gcp_permissions and gcp_predefined_role are mutually exclusive; use gcp_permissions for fine-grained custom permissions or gcp_predefined_role for a Google-managed role, not both", a.Name)
	}

	return nil
}
