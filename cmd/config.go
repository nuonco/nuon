package cmd

import (
	"github.com/powertoolsdev/go-common/config"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// install defaults
	// TODO(jm): change bucket and bucket region to use the installations prefix once we also pass in the sandbox
	// bucket nmae instead of hardcoding it.
	config.RegisterDefault("install.bucket", "nuon-installations-stage")
	config.RegisterDefault("install.bucket_region", "us-west-2")
	config.RegisterDefault("install.nuon_access_role_arn", "arn:aws:iam::618886478608:role/install-k8s-admin-stage")
	config.RegisterDefault("install.token_secret_namespace", "default")
	config.RegisterDefault("install.org_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("install.installation_state_bucket", "nuon-installations-stage")
	config.RegisterDefault("install.installation_state_bucket_region", "us-west-2")
	config.RegisterDefault("install.sandbox_bucket", "nuon-sandboxes")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// NOTE: these webhook urls are scoped at the project level, but are workflow specific. This is because we
	// create a slack notifier object at the cmd level and pass it to each individual workflow
	InstallationBotsSlackWebhookURL string `config:"installation_bots_slack_webhook_url"`
	OrgBotsSlackWebhookURL          string `config:"org_bots_slack_webhook_url"`

	// Domain specific configs
	WorkersCfg workers.Config `config:"install"`
}
