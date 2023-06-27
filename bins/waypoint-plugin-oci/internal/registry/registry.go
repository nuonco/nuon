package registry

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

var _ component.Registry = (*Registry)(nil)
var _ component.RegistryAccess = (*Registry)(nil)

func New(v *validator.Validate) (*Registry, error) {
	return &Registry{
		v: v,
	}, nil
}

type Registry struct {
	v *validator.Validate

	config configs.OCIRegistry
}
