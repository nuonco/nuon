package builder

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

var _ component.Builder = (*Builder)(nil)

type Builder struct {
	v      *validator.Validate
	config configs.NoopBuild
}

func New(v *validator.Validate) (*Builder, error) {
	return &Builder{
		v: v,
	}, nil
}
