package state

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

type Workflows struct {
	v *validator.Validate
}

type Params struct {
	fx.In

	V *validator.Validate
}

func New(params Params) (*Workflows, error) {
	return &Workflows{
		v: params.V,
	}, nil
}
