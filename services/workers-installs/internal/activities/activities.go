package activities

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/workers-installs/internal"
)

type Activities struct {
	v   *validator.Validate
	cfg *internal.Config
}

func NewActivities(v *validator.Validate,
	cfg *internal.Config) *Activities {
	return &Activities{
		v:   v,
		cfg: cfg,
	}
}
