package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type PoliciesConfig struct {
	Policies []AppPolicy `mapstructure:"policy,omitempty"`
}

func (a PoliciesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "policy", "List of policies to enforce.")
}

func (a *PoliciesConfig) parse(context.Context) error {
	return nil
}

type AppPolicy struct {
	Type     string `mapstructure:"type"`
	Contents string `mapstructure:"contents" features:"get,template"`
}

func (a AppPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "type", "Policy type which controls how it is enforced.")
	addDescription(schema, "contents", "The policy contents. Supports any reference via https://github.com/hashicorp/go-getter.")
}

func (a *AppPolicy) parse(ctx context.Context) error {
	return nil
}
