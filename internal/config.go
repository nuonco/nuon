package install

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// install defaults
	config.RegisterDefault("nuon_access_role_arn", "arn:aws:iam::618886478608:role/install-k8s-admin-stage")
	config.RegisterDefault("token_secret_namespace", "default")
	config.RegisterDefault("org_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("sandbox_bucket", "nuon-sandboxes")
}

// Config exposes a set of configuration options for the install domain
type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	HostIP string `config:"host_ip" validate:"required"`

	// NOTE: these webhook urls are scoped at the project level, but are workflow specific. This is because we
	// create a slack notifier object at the cmd level and pass it to each individual workflow
	InstallationBotsSlackWebhookURL string `config:"installation_bots_slack_webhook_url"`
	OrgBotsSlackWebhookURL          string `config:"org_bots_slack_webhook_url"`

	// NuonAccessRoleArn is the role that we add to the sandbox EKS allowed roles so we can do other operations
	// against it
	NuonAccessRoleArn string `config:"nuon_access_role_arn" validate:"required"`

	// TODO(jm): update these values to use the correct orgs account cluster values
	TokenSecretNamespace string `config:"token_secret_namespace" validate:"required"`
	OrgServerRootDomain  string `config:"org_server_root_domain" validate:"required"`

	// org IAM role template names
	OrgInstanceRoleTemplate      string `config:"orgs_instance_role_template" validate:"required"`
	OrgInstallationsRoleTemplate string `config:"orgs_installations_role_template" validate:"required"`
	OrgInstallerRoleTemplate     string `config:"orgs_installer_role_template" validate:"required"`

	InstallationsBucket       string `config:"installations_bucket" validate:"required"`
	InstallationsBucketRegion string `config:"installations_bucket_region" validate:"required"`
	SandboxBucket             string `config:"sandbox_bucket" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
