package workers

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")
	config.RegisterDefault("waypoint_token_namespace", "default")
	config.RegisterDefault("waypoint_server_root_domain", "orgs-stage.nuon.co")
}

type Config struct {
	config.Base `config:",squash"`

	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	OrgsEcrAccessRoleArn string `config:"orgs_ecr_access_role_arn" validate:"required" json:"orgs_ecr_access_iam_role_arn"`

	WaypointTokenNamespace   string `config:"waypoint_token_namespace" json:"waypoint_token_namespace" validate:"required"`
	WaypointServerRootDomain string `config:"waypoint_server_root_domain" json:"waypoint_server_root_domain" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
