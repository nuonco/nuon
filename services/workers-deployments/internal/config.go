package deployment

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "deployments")
	config.RegisterDefault("instances_temporal_namespace", "instances")
}

type Config struct {
	worker.Config `config:",squash"`

	InstancesTemporalNamespace string `config:"instances_temporal_namespace" validate:"required"`
	// waypoint configuration
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointTokenSecretTemplate  string `config:"waypoint_token_secret_template" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`

	DeploymentsBucket string `config:"deployments_bucket" validate:"required"`
	// org IAM role template names
	OrgsDeploymentsRoleTemplate string `config:"orgs_deployments_role_template" validate:"required"`

	// configuration for plans
	OrgsECRRegistryID  string `config:"orgs_ecr_registry_id" validate:"required"`
	OrgsECRRegistryARN string `config:"orgs_ecr_registry_arn" validate:"required"`
	OrgsECRRegion      string `config:"orgs_ecr_region" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
