package deprovision

import (
	"github.com/go-playground/validator/v10"

	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
)

// Activities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
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
