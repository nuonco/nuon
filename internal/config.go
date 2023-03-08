package deployment

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "deployments")
	config.RegisterDefault("temporal_task_queue", "deployment")
	config.RegisterDefault("temporal_max_concurrent_activities", 1)

	// deployment defaults
	config.RegisterDefault("waypoint_token_secret_namespace", "default")
	config.RegisterDefault("waypoint_org_server_root_domain", "orgs-stage.nuon.co")
}

type Config struct {
	config.Base `config:",squash"`

	// NOTE(jm): this should eventually be migrated to it's own temporal config in a shared package
	TemporalHost                    string `config:"temporal_host" validate:"required"`
	TemporalNamespace               string `config:"temporal_namespace" validate:"required"`
	TemporalTaskQueue               string `config:"temporal_task_queue" validate:"required"`
	TemporalMaxConcurrentActivities int    `config:"temporal_max_concurrent_activities" validate:"required"`
	HostIP                          string `config:"host_ip" validate:"required"`

	DeploymentsBucket string `config:"deployments_bucket" validate:"required"`

	// waypoint configuration
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointTokenSecretTemplate  string `config:"waypoint_token_secret_template" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`

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
