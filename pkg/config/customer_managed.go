package config

import "github.com/invopop/jsonschema"

type CustomerManagedConfig struct {
	RunnerImageURL string                    `mapstructure:"runner_image_url" toml:"runner_image_url" json:"runner_image_url" jsonschema:"required"`
	RunnerImageTag string                    `mapstructure:"runner_image_tag" toml:"runner_image_tag" json:"runner_image_tag" jsonschema:"required"`
	Platforms      []CustomerManagedPlatform `mapstructure:"platforms" toml:"platforms" json:"platforms" jsonschema:"required"`
}

func (c CustomerManagedConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("runner_image_url").Short("runner image repository").Required().
		Long("OCI repository containing the runner image included in customer-managed packages").
		Example("public.ecr.aws/example/runner").
		Field("runner_image_tag").Short("runner image tag").Required().
		Long("Tag of the runner image included in customer-managed packages").
		Example("v1.2.3").
		Field("platforms").Short("platform-specific runtime artifacts").Required().
		Long("Portal and runner binaries to include for each supported target platform")
}

type CustomerManagedPlatform struct {
	Target          string `mapstructure:"target" toml:"target" json:"target" jsonschema:"required"`
	PortalBinaryURL string `mapstructure:"portal_binary_url" toml:"portal_binary_url" json:"portal_binary_url" jsonschema:"required"`
	RunnerBinaryURL string `mapstructure:"runner_binary_url" toml:"runner_binary_url" json:"runner_binary_url" jsonschema:"required"`
}

func (p CustomerManagedPlatform) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("target").Short("target platform").Required().
		Long("Operating system and architecture for these runtime artifacts").
		Example("linux/amd64").
		Field("portal_binary_url").Short("portal binary URL").Required().
		Long("URL of the portal binary built for this target platform").
		Field("runner_binary_url").Short("runner binary URL").Required().
		Long("URL of the runner binary built for this target platform")
}
