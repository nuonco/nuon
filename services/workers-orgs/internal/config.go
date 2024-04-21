package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "orgs")
	config.RegisterDefault("waypoint_chart_dir", "/charts/waypoint")
}

type Config struct {
	worker.Config `config:",squash"`

	// NOTE: top level configuration for starting the worker
	BotsSlackWebhookURL string `config:"bots_slack_webhook_url"`
	TemporalHost        string `config:"temporal_host" validate:"required"`
	TemporalNamespace   string `config:"temporal_namespace" validate:"required"`

	WaypointServerRootDomain        string `config:"waypoint_server_root_domain" validate:"required" json:"waypoint_server_root_domain"`
	WaypointBootstrapTokenNamespace string `config:"waypoint_bootstrap_token_namespace" validate:"required" json:"waypoint_bootstrap_token_namespace"`

	// IAM roles used to manage things in orgs account
	OrgsIAMAccessRoleArn    string `config:"orgs_iam_access_role_arn" validate:"required"`
	OrgsBucketAccessRoleArn string `config:"orgs_bucket_access_role_arn" validate:"required"`
	OrgsKMSAccessRoleArn    string `config:"orgs_kms_access_role_arn" validate:"required"`

	// configs needed to access the orgs cluster, which runs all org runner/servers
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id" json:"orgs_k8s_cluster_id" validate:"required"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn" json:"orgs_k8s_role_arn" validate:"required"`
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data" json:"orgs_k8s_ca_data" validate:"required"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint" json:"orgs_k8s_public_endpoint" validate:"required"`

	// configs needed for setting up permissions to org resources
	OrgInstallationsBucketName string `config:"org_installations_bucket_name" json:"org_installations_bucket_name" validate:"required"`
	OrgDeploymentsBucketName   string `config:"org_deployments_bucket_name" json:"org_deployments_bucket_name" validate:"required"`
	OrgSecretsBucketName       string `config:"org_secrets_bucket_name" json:"org_key_values_bucket_name" validate:"required"`
	OrgsBucketName             string `config:"org_orgs_bucket_name" json:"org_orgs_bucket_name" validate:"required"`

	// configs needed to grant the workers ability to assume org iam roles
	WorkersIAMRoleARNPrefix string `config:"workers_iam_role_arn_prefix" validate:"required"`
	SupportIAMRoleARN       string `config:"support_iam_role_arn" validate:"required"`
	OrgsAccountRootARN      string `config:"orgs_account_root_arn" json:"orgs_account_root_arn"`
	OrgsAccountID           string `config:"orgs_account_id" json:"orgs_account_id"`

	// configs needed to create an IAM role for the ODR runner in the orgs account
	OrgsIAMOidcProviderURL string `config:"orgs_iam_oidc_provider_url" validate:"required"`
	OrgsIAMOidcProviderArn string `config:"orgs_iam_oidc_provider_arn" validate:"required"`
	OrgsECRRegistryArn     string `config:"orgs_ecr_registry_arn" validate:"required"`

	// configs for accessing the sandbox buckets
	SandboxBucketARN string `config:"sandbox_bucket_arn" validate:"required"`
	SandboxKeyARN    string `config:"sandbox_key_arn" validate:"required"`

	// We embed the waypoint chart locally, and use it from main here.
	WaypointChartDir string `config:"waypoint_chart_dir"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
