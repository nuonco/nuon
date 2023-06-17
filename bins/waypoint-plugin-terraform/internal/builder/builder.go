package builder

import (
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"oras.land/oras-go/v2/content/file"
)

var _ component.Builder = (*Builder)(nil)

//var _ component.BuilderODR = (*Builder)(nil)

type Builder struct {
	v      *validator.Validate
	config BuildConfig
	Store  *file.Store `validate:"required"`
}

func New(v *validator.Validate, store *file.Store) (*Builder, error) {
	return &Builder{
		Store: store,
		v:     v,
	}, nil
}
