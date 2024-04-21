package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "installs")
	config.RegisterDefault("waypoint_chart_dir", "/charts/waypoint")
}

// Config exposes a set of configuration options for the install domain
type Config struct {
	worker.Config `config:",squash"`

	// NOTE: these webhook urls are scoped at the project level, but are workflow specific. This is because we
	// create a slack notifier object at the cmd level and pass it to each individual workflow
	InstallationBotsSlackWebhookURL string `config:"installation_bots_slack_webhook_url" validate:"required"`
	OrgBotsSlackWebhookURL          string `config:"org_bots_slack_webhook_url" validate:"required"`

	// NuonAccessRoleArn is the role that we add to the sandbox EKS allowed roles so we can do other operations
	// against it
	NuonAccessRoleArn string `config:"nuon_access_role_arn" validate:"required"`

	// TODO(jm): update these values to use the correct orgs account cluster values
	TokenSecretNamespace string `config:"token_secret_namespace" validate:"required"`
	OrgServerRootDomain  string `config:"org_server_root_domain" validate:"required"`

	// org IAM role template names
	OrgInstanceRoleTemplate      string `config:"orgs_instances_role_template" validate:"required"`
	OrgInstallationsRoleTemplate string `config:"orgs_installations_role_template" validate:"required"`
	OrgInstallerRoleTemplate     string `config:"orgs_installer_role_template" validate:"required"`

	InstallationsBucket       string `config:"installations_bucket" validate:"required"`
	InstallationsBucketRegion string `config:"installations_bucket_region" validate:"required"`
	SandboxBucket             string `config:"sandbox_bucket" validate:"required"`

	// authenticate with orgs cluster
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data" validate:"required"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint" validate:"required"`
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id" validate:"required"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn" validate:"required"`

	// PublicDomain configuration
	PublicDomain           string `config:"public_domain" validate:"required"`
	PublicDomainZoneID     string `config:"public_domain_zone_id" validate:"required"`
	PublicDNSAccessRoleARN string `config:"public_dns_access_role_arn" validate:"required"`

	// We embed the waypoint chart locally, and use it from main here.
	WaypointChartDir string `config:"waypoint_chart_dir" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
