package platform

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

var _ component.Platform = (*Platform)(nil)
var _ component.Destroyer = (*Platform)(nil)

func New(v *validator.Validate) (*Platform, error) {
	return &Platform{
		v: v,
	}, nil
}

type Platform struct {
	v *validator.Validate

	// internal fields
	Cfg configs.JobDeploy `validate:"required"`
}
