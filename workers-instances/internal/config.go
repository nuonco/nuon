package instance

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "instances")

	// instance defaults
	config.RegisterDefault("waypoint_token_secret_namespace", "default")
	config.RegisterDefault("waypoint_server_root_domain", "orgs-stage.nuon.co")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	DeploymentBotsSlackWebhookURL string `config:"deployment_bots_slack_webhook_url" validate:"required"`
	DeploymentsBucket             string `config:"deployments_bucket" validate:"required"`

	// waypoint configuration
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`

	// org IAM role template names
	OrgsDeploymentsRoleTemplate string `config:"orgs_deployments_role_template" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
