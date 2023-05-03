package instance

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "instances")
}

type Config struct {
	worker.Config `config:",squash"`

	DeploymentBotsSlackWebhookURL string `config:"deployment_bots_slack_webhook_url" validate:"required"`
	DeploymentsBucket             string `config:"deployments_bucket" validate:"required"`

	// waypoint configuration
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`

	// org IAM role template names
	OrgsDeploymentsRoleTemplate string `config:"orgs_deployments_role_template" validate:"required"`

	// authenticate with orgs cluster
	OrgsK8sCAData         string `config:"orgs_k8s_ca_data"`
	OrgsK8sPublicEndpoint string `config:"orgs_k8s_public_endpoint"`
	OrgsK8sClusterID      string `config:"orgs_k8s_cluster_id"`
	OrgsK8sRoleArn        string `config:"orgs_k8s_role_arn"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
