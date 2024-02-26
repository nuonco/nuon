package sync

import (
	"github.com/go-playground/validator/v10"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
)

type Activities struct {
	v   *validator.Validate
	cfg workers.Config
}

func NewActivities(v *validator.Validate, cfg workers.Config) *Activities {
	return &Activities{
		v:   v,
		cfg: cfg,
	}
}
