package config

import (
	"github.com/invopop/jsonschema"
)

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty" toml:"iam_role_arn,omitempty" jsonschema:"required"`
	AWSRegion  string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
	ImageURL   string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required"`
	Tag        string `mapstructure:"tag,omitempty" toml:"tag,omitempty" jsonschema:"required"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required" `
	Tag      string `mapstructure:"tag,omitempty" toml:"tag,omitempty" jsonschema:"required"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type GCPGARConfig struct {
	GCPProjectID             string `mapstructure:"gcp_project_id,omitempty" toml:"gcp_project_id,omitempty" jsonschema:"required"`
	GCPRegion                string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
	ImageURL                 string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required"`
	Tag                      string `mapstructure:"tag,omitempty" toml:"tag,omitempty" jsonschema:"required"`
	ServiceAccountEmail      string `mapstructure:"service_account_email,omitempty" toml:"service_account_email,omitempty"`
	WorkloadIdentityProvider string `mapstructure:"workload_identity_provider,omitempty" toml:"workload_identity_provider,omitempty"`
}

type ExternalImageComponentConfig struct {
	AWSECRImageConfig *AWSECRConfig      `mapstructure:"aws_ecr,omitempty" toml:"aws_ecr,omitempty" jsonschema:"oneof_required=ecr_source"`
	GCPGARImageConfig *GCPGARConfig      `mapstructure:"gcp_gar,omitempty" toml:"gcp_gar,omitempty" jsonschema:"oneof_required=gar_source"`
	PublicImageConfig *PublicImageConfig `mapstructure:"public,omitempty" toml:"public,omitempty" jsonschema:"oneof_required=public_source"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`
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

func (g GCPGARConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("gcp_project_id").Short("GCP project ID").Required().
		Long("Google Cloud project ID where the Artifact Registry repository is located").
		Example("my-gcp-project").
		Field("region").Short("GCP region for the GAR repository").Required().
		Long("Google Cloud region where the Artifact Registry repository is located").
		Example("us-central1").
		Example("us-east1").
		Example("europe-west1").
		Field("image_url").Short("GAR image URL").Required().
		Long("Full URL to the GAR image (without tag). Format: <region>-docker.pkg.dev/<project>/<repository>/<image>").
		Example("us-central1-docker.pkg.dev/my-project/my-repo/my-image").
		Field("tag").Short("image tag").Required().
		Long("Tag or version of the container image to deploy. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}").
		Field("service_account_email").Short("GCP service account for impersonation").
		Long("Optional service account email to impersonate when pulling from GAR. If not set, uses application default credentials").
		Example("my-sa@my-project.iam.gserviceaccount.com")
}

func (e ExternalImageComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("aws_ecr").Short("AWS ECR image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from AWS Elastic Container Registry. Use when deploying images from private ECR repositories").
		Field("gcp_gar").Short("GCP Artifact Registry image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from Google Artifact Registry. Use when deploying images from private GAR repositories").
		Field("public").Short("public registry image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from public container registries (Docker Hub, Quay.io, GCR, etc)").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\"). Default: 15m. Max: 1h").
		Default("15m").
		Example("30m").
		Example("1h").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\"). Default: 5m. Max: 1h").
		Default("5m").
		Example("30m").
		Example("1h")
}

func (t *ExternalImageComponentConfig) Validate() error {
	return nil
}

func (t *ExternalImageComponentConfig) Parse() error {
	return nil
}
