package installdelegationdns

import (
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Activities struct {
	v   *validator.Validate
	cfg *internal.Config
}

func NewActivities(v *validator.Validate, cfg *internal.Config) *Activities {
	return &Activities{
		v:   v,
		cfg: cfg,
	}
}
