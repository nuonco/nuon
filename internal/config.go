package deployment

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Bucket       string `config:"bucket" validate:"required"`
	BucketRegion string `config:"bucket_region" validate:"required"`
	RoleArn      string `config:"role_arn" validate:"required"`

	TemporalHost      string `config:"temporal_host" validate:"required"`
	TemporalNamespace string `config:"temporal_namespace" validate:"required"`

	WaypointTokenSecretNamespace string `config:"waypoint_token_secret_namespace" validate:"required"`
	WaypointOrgServerRootDomain  string `config:"waypoint_org_server_root_domain" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
