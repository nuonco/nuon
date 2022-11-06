package cmd

import (
	"github.com/powertoolsdev/go-common/config"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// deployment defaults
	config.RegisterDefault("deployment.waypoint_token_secret_namespace", "default")
	config.RegisterDefault("deployment.bucket", "nuon-installations-stage")
	config.RegisterDefault("deployment.bucket_region", "us-west-2")
	config.RegisterDefault("deployment.role_arn", "arn:aws:iam::618886478608:role/install-k8s-admin-stage")
	config.RegisterDefault("deployment.waypoint_org_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("deployment.temporal_host", "localhost:7233")
	config.RegisterDefault("deployment.temporal_namespace", "default")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// NOTE: these webhook urls are scoped at the project level, but are workflow specific. This is because we
	// create a slack notifier object at the cmd level and pass it to each individual workflow
	InstallationBotsSlackWebhookURL string `config:"installation_bots_slack_webhook_url"`

	// Domain specific configs
	DeploymentCfg workers.Config `config:"deployment"`
}
