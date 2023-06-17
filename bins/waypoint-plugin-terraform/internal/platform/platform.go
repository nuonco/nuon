package platform

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
	"oras.land/oras-go/v2/content/file"
)

var _ component.Platform = (*Platform)(nil)
var _ component.Destroyer = (*Platform)(nil)

func New(v *validator.Validate, store *file.Store) (*Platform, error) {
	return &Platform{
		v:     v,
		Store: store,
	}, nil
}

type Platform struct {
	v *validator.Validate

	// internal fields
	Cfg       configs.TerraformDeploy `validate:"required"`
	Workspace workspace.Workspace     `validate:"required"`
	Store     *file.Store             `validate:"required"`
}
