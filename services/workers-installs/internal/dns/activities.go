package dns

import (
	"github.com/go-playground/validator/v10"
)

// Activities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type Activities struct {
	v *validator.Validate
}

func NewActivities(v *validator.Validate) *Activities {
	return &Activities{
		v: v,
	}
}
