package registry

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"oras.land/oras-go/v2/content/file"
)

var _ component.Registry = (*Registry)(nil)
var _ component.RegistryAccess = (*Registry)(nil)

func New(v *validator.Validate, store *file.Store) (*Registry, error) {
	return &Registry{
		v:     v,
		Store: store,
	}, nil
}

type Registry struct {
	v *validator.Validate

	config configs.TerraformBuildAWSECRRegistry
	Store  *file.Store
}
