package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type AppAWSIAMPolicy struct {
	ManagedPolicyName string `mapstructure:"managed_policy_name,omitempty"`
	Name              string `mapstructure:"name" jsonschema:"required"`

	Contents string `mapstructure:"contents" features:"template,get"`
}

func (a AppAWSIAMPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "managed_policy_name", "policy name to use for an AWS managed policy.")
	addDescription(schema, "name", "policy name to use for a policy. Supports Nuon templates.")
	addDescription(schema, "contents", "json contents to use for the policy. Supports loading policies via https://github.com/hashicorp/go-getter.")
}

func (a *AppAWSIAMPolicy) parse(ctx context.Context) error {
	return nil
}
