package execute

import (
	"github.com/go-playground/validator/v10"
)

type Activities struct {
	v *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		v: validator.New(),
	}
}
