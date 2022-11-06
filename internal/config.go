package instance

import "github.com/go-playground/validator/v10"

type Config struct {
	Bucket       string `config:"bucket" validate:"required"`
	BucketRegion string `config:"bucket_region" validate:"required"`
	RoleArn      string `config:"role_arn" validate:"required"`

	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointServerRootDomain     string `config:"waypoint_server_root_domain" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
