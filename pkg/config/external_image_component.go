package config

import (
	"github.com/invopop/jsonschema"
)

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty" jsonschema:"required"`
	AWSRegion  string `mapstructure:"region,omitempty" jsonschema:"required"`
	ImageURL   string `mapstructure:"image_url,omitempty" jsonschema:"required"`
	Tag        string `mapstructure:"tag,omitempty" jsonschema:"required"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url,omitempty" jsonschema:"required" `
	Tag      string `mapstructure:"tag,omitempty" jsonschema:"required"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type ExternalImageComponentConfig struct {
	AWSECRImageConfig *AWSECRConfig      `mapstructure:"aws_ecr,omitempty" jsonschema:"oneof_required=public"`
	PublicImageConfig *PublicImageConfig `mapstructure:"public,omitempty" jsonschema:"oneof_required=aws_ecr"`
}

func (a AWSECRConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("iam_role_arn").Short("IAM role ARN for ECR access").Required().
		Long("ARN of the IAM role with permissions to pull images from the ECR repository").
		Example("arn:aws:iam::123456789012:role/ecr-pull-role").
		Field("region").Short("AWS region for the ECR repository").Required().
		Long("AWS region where the ECR repository is located").
		Example("us-east-1").
		Example("us-west-2").
		Example("eu-west-1").
		Field("image_url").Short("ECR image URL").Required().
		Long("Full URL to the ECR image (without tag). Format: <account-id>.dkr.ecr.<region>.amazonaws.com/<repository-name>/<image-name>").
		Example("123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp/api").
		Example("123456789012.dkr.ecr.us-west-2.amazonaws.com/myapp/worker").
		Field("tag").Short("image tag").Required().
		Long("Tag or version of the container image to deploy. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}")
}

func (p PublicImageConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("image_url").Short("container image URL").Required().
		Long("Full URL to the container image from a public registry (Docker Hub, Quay.io, etc). Format: [registry/]<repository>/<image-name>").
		Example("nginx:latest").
		Example("docker.io/library/postgres").
		Example("quay.io/myorg/myapp").
		Example("gcr.io/myproject/myapp").
		Field("tag").Short("image tag").Required().
		Long("Tag or version of the container image to deploy. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}")
}

func (e ExternalImageComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("source path or URL").
		Long("Optional source path or URL for the component configuration. Supports HTTP(S) URLs, git repositories, file paths, and relative paths (./). Examples: https://example.com/config.yaml, git::https://github.com/org/repo//config.yaml, file:///path/to/config.yaml, ./local/config.yaml").
		Field("type").Short("component type").
		Field("name").Short("component name").
		Field("var_name").Short("variable name for component output").
		Long("Optional name to use when storing component outputs as variables. If not specified, uses the component name").
		Field("dependencies").Short("component dependencies").
		Long("List of other components that must be deployed before this component. Automatically extracted from template references").
		Field("aws_ecr").Short("AWS ECR image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from AWS Elastic Container Registry. Use when deploying images from private ECR repositories").
		Field("public").Short("public registry image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from public container registries (Docker Hub, Quay.io, GCR, etc)")
}

func (t *ExternalImageComponentConfig) Validate() error {
	return nil
}

func (t *ExternalImageComponentConfig) Parse() error {
	return nil
}
