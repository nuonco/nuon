package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type AppAWSIAMPolicy struct {
	ManagedPolicyName string `mapstructure:"managed_policy_name,omitempty" toml:"managed_policy_name,omitempty"`
	Name              string `mapstructure:"name" toml:"name" jsonschema:"required"`

	Contents string `mapstructure:"contents" toml:"contents" features:"template,get"`
}

func (a AppAWSIAMPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("managed_policy_name").Short("AWS managed policy name").
		Long("Name or ARN of an AWS managed policy to attach to the role").
		Example("AmazonS3FullAccess").
		Example("ReadOnlyAccess").
		Field("name").Short("policy name").Required().
		Long("Name for the policy. Used when creating inline or customer managed policies. Supports Nuon templating").
		Example("app-{{.nuon.install.id}}-policy").
		Example("s3-access-policy").
		Field("contents").Short("IAM policy document").
		Long("JSON policy document defining IAM permissions. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/policy.json), git repositories (git::https://github.com/org/repo//policy.json), file paths (file:///path/to/policy.json), and relative paths (./policy.json)").
		Example("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"s3:*\",\"Resource\":\"*\"}]}")
}

func (a *AppAWSIAMPolicy) parse(ctx context.Context) error {
	return nil
}
