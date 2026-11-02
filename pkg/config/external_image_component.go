package config

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/pkg/oci/updatepolicy"
)

type AWSECRConfig struct {
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty" toml:"iam_role_arn,omitempty"`
	AWSRegion  string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
	ImageURL   string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required"`
	Tag        string `mapstructure:"tag,omitempty" toml:"tag,omitempty"`
	// UpdatePolicy is an optional Masterminds-compatible semver constraint
	// (e.g. "~1.25.0", "^2"). When set, the runner picks the highest
	// matching tag from the registry at build time. Either tag or
	// update_policy must be set.
	UpdatePolicy string `mapstructure:"update_policy,omitempty" toml:"update_policy,omitempty"`
}

type PublicImageConfig struct {
	ImageURL string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required" `
	Tag      string `mapstructure:"tag,omitempty" toml:"tag,omitempty"`
	// UpdatePolicy is an optional Masterminds-compatible semver constraint
	// (e.g. "~1.25.0", "^2"). When set, the runner picks the highest
	// matching tag from the registry at build time. Either tag or
	// update_policy must be set.
	UpdatePolicy string `mapstructure:"update_policy,omitempty" toml:"update_policy,omitempty"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type GCPGARConfig struct {
	UseCloudConnection       bool   `mapstructure:"use_cloud_connection,omitempty" toml:"use_cloud_connection,omitempty" nuonhash:"omitempty"`
	GCPProjectID             string `mapstructure:"gcp_project_id,omitempty" toml:"gcp_project_id,omitempty" jsonschema:"required"`
	GCPRegion                string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
	ImageURL                 string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required"`
	Tag                      string `mapstructure:"tag,omitempty" toml:"tag,omitempty"`
	ServiceAccountEmail      string `mapstructure:"service_account_email,omitempty" toml:"service_account_email,omitempty"`
	WorkloadIdentityProvider string `mapstructure:"workload_identity_provider,omitempty" toml:"workload_identity_provider,omitempty"`
	// UpdatePolicy is an optional Masterminds-compatible semver constraint
	// (e.g. "~1.25.0", "^2"). When set, the runner picks the highest
	// matching tag from the registry at build time. Either tag or
	// update_policy must be set.
	UpdatePolicy string `mapstructure:"update_policy,omitempty" toml:"update_policy,omitempty"`
}

type AzureACRConfig struct {
	UseCloudConnection bool   `mapstructure:"use_cloud_connection,omitempty" toml:"use_cloud_connection,omitempty" nuonhash:"omitempty"`
	ImageURL           string `mapstructure:"image_url,omitempty" toml:"image_url,omitempty" jsonschema:"required"`
	Tag                string `mapstructure:"tag,omitempty" toml:"tag,omitempty"`
	RegistryURL        string `mapstructure:"registry_url,omitempty" toml:"registry_url,omitempty" jsonschema:"required"`
	TenantID           string `mapstructure:"tenant_id,omitempty" toml:"tenant_id,omitempty"`
	ClientID           string `mapstructure:"client_id,omitempty" toml:"client_id,omitempty"`
	// These name an AppSecret holding the credential; the "_name" suffix is
	// load-bearing, because a field called client_secret invites pasting the
	// secret itself into a file that gets committed.
	ClientSecretName      string `mapstructure:"client_secret_name,omitempty" toml:"client_secret_name,omitempty"`
	ClientCertificateName string `mapstructure:"client_certificate_name,omitempty" toml:"client_certificate_name,omitempty"`
	// UpdatePolicy is an optional Masterminds-compatible semver constraint
	// (e.g. "~1.25.0", "^2"). When set, the runner picks the highest
	// matching tag from the registry at build time. Either tag or
	// update_policy must be set.
	UpdatePolicy string `mapstructure:"update_policy,omitempty" toml:"update_policy,omitempty"`
}

type ExternalImageComponentConfig struct {
	AWSECRImageConfig   *AWSECRConfig           `mapstructure:"aws_ecr,omitempty" toml:"aws_ecr,omitempty" jsonschema:"oneof_required=ecr_source"`
	GCPGARImageConfig   *GCPGARConfig           `mapstructure:"gcp_gar,omitempty" toml:"gcp_gar,omitempty" jsonschema:"oneof_required=gar_source"`
	AzureACRImageConfig *AzureACRConfig         `mapstructure:"azure_acr,omitempty" toml:"azure_acr,omitempty" jsonschema:"oneof_required=acr_source"`
	PublicImageConfig   *PublicImageConfig      `mapstructure:"public,omitempty" toml:"public,omitempty" jsonschema:"oneof_required=public_source"`
	Verification        *signature.Verification `mapstructure:"verification,omitempty" toml:"verification,omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`
}

func (a AWSECRConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("iam_role_arn").Short("ECR pull role ARN").
		Long("Assume this role for legacy ECR authentication. Omit to use the verified identity saved for the exact image repository.").
		Field("region").Short("AWS region for the ECR repository").Required().
		Long("AWS region where the ECR repository is located").
		Example("us-east-1").
		Example("us-west-2").
		Example("eu-west-1").
		Field("image_url").Short("ECR image URL").Required().
		Long("Full URL to the ECR image (without tag). Format: <account-id>.dkr.ecr.<region>.amazonaws.com/<repository-name>/<image-name>").
		Example("123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp/api").
		Example("123456789012.dkr.ecr.us-west-2.amazonaws.com/myapp/worker").
		Field("tag").Short("image tag").
		Long("Tag or version of the container image to deploy. Either tag or update_policy must be set. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}").
		Field("update_policy").Short("semver constraint for tag resolution").
		Long("Semver constraint for picking a tag at build time. When set, at each build the runner lists tags from the registry, filters to those that parse as semver and satisfy the constraint, and selects the highest matching tag. Tags that aren't valid semver (e.g. \"latest\", \"stable\", branch names) are skipped. Either tag or update_policy must be set. Supported constraint shapes: tilde (~1.25.0 → >=1.25.0 <1.26.0), caret (^2.3.1 → >=2.3.1 <3.0.0), comparators joined with AND (>=1.0.0,<2.0.0), wildcard ranges (1.2.x, 1.x), inclusive hyphen ranges (1.0.0 - 2.0.0), OR (^1.0 || ^2.0), and exact match (=1.25.5).").
		Example("~1.25.0").
		Example("^2.0.0").
		Example(">=1.0.0,<2.0.0").
		Example("1.x").
		Example("^1.0 || ^2.0")
}

func (p PublicImageConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("image_url").Short("container image URL").Required().
		Long("Full URL to the container image from a public registry (Docker Hub, Quay.io, etc). Format: [registry/]<repository>/<image-name>").
		Example("nginx:latest").
		Example("docker.io/library/postgres").
		Example("quay.io/myorg/myapp").
		Example("gcr.io/myproject/myapp").
		Field("tag").Short("image tag").
		Long("Tag or version of the container image to deploy. Either tag or update_policy must be set. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}").
		Field("update_policy").Short("semver constraint for tag resolution").
		Long("Semver constraint for picking a tag at build time. When set, at each build the runner lists tags from the registry, filters to those that parse as semver and satisfy the constraint, and selects the highest matching tag. Tags that aren't valid semver (e.g. \"latest\", \"stable\", branch names) are skipped. Either tag or update_policy must be set. Supported constraint shapes: tilde (~1.25.0 → >=1.25.0 <1.26.0), caret (^2.3.1 → >=2.3.1 <3.0.0), comparators joined with AND (>=1.0.0,<2.0.0), wildcard ranges (1.2.x, 1.x), inclusive hyphen ranges (1.0.0 - 2.0.0), OR (^1.0 || ^2.0), and exact match (=1.25.5).").
		Example("~1.25.0").
		Example("^2.0.0").
		Example(">=1.0.0,<2.0.0").
		Example("1.x").
		Example("^1.0 || ^2.0")
}

func (g GCPGARConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("use_cloud_connection").Short("use the saved image repository identity").
		Long("Set true to use a verified Cloud Connection for the exact repository. Omit to preserve ambient Google authentication or legacy impersonation/federation. Cannot be combined with service_account_email or workload_identity_provider.").
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
		Field("tag").Short("image tag").
		Long("Tag or version of the container image to deploy. Either tag or update_policy must be set. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}").
		Field("update_policy").Short("semver constraint for tag resolution").
		Long("Semver constraint for picking a tag at build time. When set, at each build the runner lists tags from the registry, filters to those that parse as semver and satisfy the constraint, and selects the highest matching tag. Tags that aren't valid semver (e.g. \"latest\", \"stable\", branch names) are skipped. Either tag or update_policy must be set. Supported constraint shapes: tilde (~1.25.0 → >=1.25.0 <1.26.0), caret (^2.3.1 → >=2.3.1 <3.0.0), comparators joined with AND (>=1.0.0,<2.0.0), wildcard ranges (1.2.x, 1.x), inclusive hyphen ranges (1.0.0 - 2.0.0), OR (^1.0 || ^2.0), and exact match (=1.25.5).").
		Example("~1.25.0").
		Example("^2.0.0").
		Example(">=1.0.0,<2.0.0").
		Example("1.x").
		Example("^1.0 || ^2.0").
		Field("service_account_email").Short("legacy service account").
		Long("Service account to impersonate for legacy GAR authentication. Omit both legacy fields for ambient Google authentication, or set use_cloud_connection = true for the saved repository identity.").
		Field("workload_identity_provider").Short("legacy federation provider").
		Long("Workload identity provider for legacy federation, optionally with service_account_email for impersonation")
}

func (g GCPGARConfig) ValidateCredentials() error {
	if g.UseCloudConnection && (g.ServiceAccountEmail != "" || g.WorkloadIdentityProvider != "") {
		return fmt.Errorf("gcp_gar: use_cloud_connection cannot be combined with service_account_email or workload_identity_provider")
	}
	return nil
}

func (a AzureACRConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("use_cloud_connection").Short("use the saved image repository identity").
		Long("Set true to use a verified Cloud Connection for the exact repository. Omit to preserve ambient Azure authentication or explicit secret/certificate credentials. Cannot be combined with component credentials.").
		Field("image_url").Short("ACR image URL").Required().
		Long("Full URL to the ACR image (without tag). Format: <registry>.azurecr.io/<repository>/<image>").
		Example("myregistry.azurecr.io/myapp/api").
		Field("registry_url").Short("ACR login server").Required().
		Long("Azure Container Registry login server. Format: <registry>.azurecr.io").
		Example("myregistry.azurecr.io").
		Field("tag").Short("image tag").
		Long("Tag or version of the container image to deploy. Either tag or update_policy must be set. Supports templating (e.g., {{.nuon.install.id}})").
		Example("v1.0.0").
		Example("latest").
		Example("{{.nuon.install.id}}").
		Field("update_policy").Short("semver constraint for tag resolution").
		Long("Semver constraint for picking a tag at build time. When set, at each build the runner lists tags from the registry, filters to those that parse as semver and satisfy the constraint, and selects the highest matching tag. Tags that aren't valid semver (e.g. \"latest\", \"stable\", branch names) are skipped. Either tag or update_policy must be set. Supported constraint shapes: tilde (~1.25.0 → >=1.25.0 <1.26.0), caret (^2.3.1 → >=2.3.1 <3.0.0), comparators joined with AND (>=1.0.0,<2.0.0), wildcard ranges (1.2.x, 1.x), inclusive hyphen ranges (1.0.0 - 2.0.0), OR (^1.0 || ^2.0), and exact match (=1.25.5).").
		Example("~1.25.0").
		Example("^2.0.0").
		Example(">=1.0.0,<2.0.0").
		Example("1.x").
		Example("^1.0 || ^2.0").
		Field("tenant_id").Short("tenant ID (secret/certificate mode only)").
		Long("Omit for OIDC. Otherwise set tenant_id, client_id, and exactly one of client_secret_name or client_certificate_name.").
		Example("00000000-0000-0000-0000-000000000000").
		Field("client_id").Short("application client ID (secret/certificate mode only)").
		Long("Omit for OIDC. For explicit credentials, use the application client ID in tenant_id with the required repository permissions.").
		Example("00000000-0000-0000-0000-000000000000").
		Field("client_secret_name").Short("app secret name (secret authentication only)").
		Long("Omit for OIDC. Name of the app secret holding the client secret, not its value. Set with tenant_id and client_id; do not set client_certificate_name.").
		Example("azure-acr-client-secret").
		Field("client_certificate_name").Short("app secret name (certificate authentication only)").
		Long("Omit for OIDC. App secret name for a base64-encoded PEM private key and certificate. Set tenant_id and client_id; omit client_secret_name.").
		Example("azure-acr-client-cert")
}

func (a AzureACRConfig) ValidateCredentials() error {
	anySet := a.TenantID != "" || a.ClientID != "" || a.ClientSecretName != "" || a.ClientCertificateName != ""
	if a.UseCloudConnection && anySet {
		return fmt.Errorf("azure_acr: use_cloud_connection cannot be combined with component-level credentials")
	}
	if a.ClientSecretName != "" && a.ClientCertificateName != "" {
		return fmt.Errorf("azure_acr: only one of client_secret_name or client_certificate_name may be set")
	}

	if !anySet {
		return nil
	}

	var missing []string
	if a.TenantID == "" {
		missing = append(missing, "tenant_id")
	}
	if a.ClientID == "" {
		missing = append(missing, "client_id")
	}
	if a.ClientSecretName == "" && a.ClientCertificateName == "" {
		missing = append(missing, "one of client_secret_name or client_certificate_name")
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"azure_acr: incomplete credentials, missing %s; set all of tenant_id, client_id and a credential, or omit credentials for ambient authentication",
			strings.Join(missing, ", "),
		)
	}

	return nil
}

func (e ExternalImageComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("aws_ecr").Short("AWS ECR image configuration").OneOfRequired("image_source").
		Long("Set iam_role_arn for legacy role authentication, or omit it to use a verified ECR identity saved for the exact repository in Settings > Image repositories.").
		Field("gcp_gar").Short("GCP Artifact Registry image configuration").OneOfRequired("image_source").
		Long("Set use_cloud_connection = true for the saved repository identity. Otherwise use ambient Google authentication or legacy service_account_email/workload_identity_provider credentials.").
		Field("azure_acr").Short("Azure Container Registry image configuration").OneOfRequired("image_source").
		Long("Set use_cloud_connection = true for the saved repository identity. Otherwise use ambient Azure authentication or explicit secret/certificate credentials.").
		Field("public").Short("public registry image configuration").OneOfRequired("image_source").
		Long("Configuration for pulling images from public container registries (Docker Hub, Quay.io, GCR, etc)").
		Field("verification").Short("container image signature verification policy").
		Long("Require the resolved image digest to satisfy at least one configured Sigstore keyless identity or Cosign public key before Nuon copies it").
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
	if err := t.Verification.Validate(); err != nil {
		return err
	}

	// Every image source must declare either a literal tag or an
	// `update_policy` semver constraint (or both); update_policy syntax is
	// validated up-front so users get a clear error before the API ever
	// rejects the sync.
	type imageSource struct {
		name         string
		tag          string
		updatePolicy string
	}
	var sources []imageSource
	if t.PublicImageConfig != nil {
		sources = append(sources, imageSource{"public", t.PublicImageConfig.Tag, t.PublicImageConfig.UpdatePolicy})
	}
	if t.AWSECRImageConfig != nil {
		sources = append(sources, imageSource{"aws_ecr", t.AWSECRImageConfig.Tag, t.AWSECRImageConfig.UpdatePolicy})
	}
	if t.GCPGARImageConfig != nil {
		if err := t.GCPGARImageConfig.ValidateCredentials(); err != nil {
			return err
		}
		sources = append(sources, imageSource{"gcp_gar", t.GCPGARImageConfig.Tag, t.GCPGARImageConfig.UpdatePolicy})
	}
	if t.AzureACRImageConfig != nil {
		sources = append(sources, imageSource{"azure_acr", t.AzureACRImageConfig.Tag, t.AzureACRImageConfig.UpdatePolicy})
		if err := t.AzureACRImageConfig.ValidateCredentials(); err != nil {
			return err
		}
	}
	for _, s := range sources {
		if s.tag == "" && s.updatePolicy == "" {
			return fmt.Errorf("%s: either tag or update_policy must be set", s.name)
		}
		if s.updatePolicy != "" {
			if err := updatepolicy.Validate(s.updatePolicy); err != nil {
				return fmt.Errorf("%s: %w", s.name, err)
			}
		}
	}
	return nil
}

func (t *ExternalImageComponentConfig) Parse() error {
	return nil
}
