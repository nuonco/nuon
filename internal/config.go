package deployment

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// deployment defaults
	config.RegisterDefault("waypoint_token_secret_namespace", "default")
	config.RegisterDefault("waypoint_org_server_root_domain", "orgs-stage.nuon.co")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	DeploymentsBucket string `config:"bucket" validate:"required"`

	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointOrgServerRootDomain  string `config:"waypoint_org_server_root_domain" validate:"required"`

	// org IAM role template names
	OrgInstanceRoleTemplate      string `config:"orgs_instance_role_template" validate:"required"`
	OrgInstallationsRoleTemplate string `config:"orgs_installations_role_template" validate:"required"`
	OrgInstallerRoleTemplate     string `config:"orgs_installer_role_template" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
