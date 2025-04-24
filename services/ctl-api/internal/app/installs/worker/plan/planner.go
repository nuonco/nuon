package plan

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

type Planner struct {
	v *validator.Validate
}

type Params struct {
	fx.In

	V *validator.Validate
}
