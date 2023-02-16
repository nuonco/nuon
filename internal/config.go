package deployment

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-config/pkg/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// deployment defaults
	config.RegisterDefault("waypoint_token_secret_namespace", "default")
	config.RegisterDefault("waypoint_org_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("waypoint_token_secret_template", "waypoint-bootstrap-token-%s")

	config.RegisterDefault("sandbox_bucket", "nuon-sandboxes")
	config.RegisterDefault("sandbox_bucket_region", "us-west-2")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host" validate:"required"`
	TemporalNamespace string `config:"temporal_namespace" validate:"required"`

	DeploymentsBucket         string `config:"deployments_bucket" validate:"required"`
	DeploymentsBucketRegion   string `config:"deployments_bucket_region" validate:"required"`
	InstallationsBucket       string `config:"installations_bucket" validate:"required"`
	InstallationsBucketRegion string `config:"installations_bucket_region" validate:"required"`
	InstancesBucket           string `config:"instances_bucket" validate:"required"`
	InstancesBucketRegion     string `config:"instances_bucket_region" validate:"required"`
	OrgsBucket                string `config:"orgs_bucket" validate:"required"`
	OrgsBucketRegion          string `config:"orgs_bucket_region" validate:"required"`

	// waypoint configuration
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`
	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointTokenSecretTemplate  string `config:"waypoint_token_secret_template" validate:"required"`

	// org IAM role template names
	OrgsDeploymentsRoleTemplate   string `config:"orgs_deployments_role_template" validate:"required"`
	OrgsInstallerRoleTemplate     string `config:"orgs_installer_role_template" validate:"required"`
	OrgsInstallationsRoleTemplate string `config:"orgs_installations_role_template" validate:"required"`
	OrgsInstancesRoleTemplate     string `config:"orgs_instances_role_template" validate:"required"`
	OrgsOdrRoleTemplate           string `config:"orgs_odr_role_template" validate:"required"`
	OrgsOrgsRoleTemplate          string `config:"orgs_orgs_role_template" validate:"required"`

	// configuration for waypoint plans
	OrgsECRRegistryID  string `config:"orgs_ecr_registry_id" validate:"required"`
	OrgsECRRegistryARN string `config:"orgs_ecr_registry_arn" validate:"required"`
	OrgsECRRegion      string `config:"orgs_ecr_region" validate:"required" faker:"oneof: us-west-2"`

	// configuration for terraform plans
	SandboxBucket       string `config:"sandbox_bucket" validate:"required"`
	SandboxBucketRegion string `config:"sandbox_bucket_region" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
