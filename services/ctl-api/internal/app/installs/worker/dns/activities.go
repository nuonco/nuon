package installdelegationdns

import (
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// Activities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
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
