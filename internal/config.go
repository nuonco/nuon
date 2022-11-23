package workers

import "github.com/go-playground/validator/v10"

type Config struct {
	OrgsEcrAccessIamRoleArn  string `config:"orgs_ecr_access_iam_role_arn" validate:"required" json:"orgs_ecr_access_iam_role_arn"`
	WaypointTokenNamespace   string `config:"waypoint_token_namespace" json:"waypoint_token_namespace" validate:"required"`
	WaypointServerRootDomain string `config:"waypoint_server_root_domain" json:"waypoint_server_root_domain" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
