package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "apps")
	config.RegisterDefault("temporal_task_queue", "apps")

	config.RegisterDefault("waypoint_token_namespace", "default")
	config.RegisterDefault("waypoint_server_root_domain", "orgs-stage.nuon.co")

	//org bucket default
	config.RegisterDefault("org_orgs_bucket_name", "nuon-orgs-stage")
}

type Config struct {
	worker.Config `config:",squash"`

	OrgsEcrAccessRoleArn string `config:"orgs_ecr_access_role_arn" validate:"required" json:"orgs_ecr_access_iam_role_arn"`
	OrgsBucketName       string `config:"org_orgs_bucket_name" json:"org_orgs_bucket_name" validate:"required"`
	OrgsRoleTemplate     string `config:"orgs_role_template" validate:"required"`

	WaypointTokenNamespace   string `config:"waypoint_token_namespace" json:"waypoint_token_namespace" validate:"required"`
	WaypointServerRootDomain string `config:"waypoint_server_root_domain" json:"waypoint_server_root_domain" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
