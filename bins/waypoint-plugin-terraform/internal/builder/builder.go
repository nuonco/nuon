package builder

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
)

var _ component.Builder = (*Builder)(nil)
var _ component.BuilderODR = (*Builder)(nil)

type Builder struct {
	v      *validator.Validate
	config BuildConfig
}

func New(v *validator.Validate) (*Builder, error) {
	return &Builder{
		v: v,
	}, nil
}
