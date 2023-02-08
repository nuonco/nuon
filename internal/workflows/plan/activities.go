package plan

import "github.com/go-playground/validator/v10"

type Activities struct {
	v           *validator.Validate
	planCreator planCreator
}

func NewActivities() *Activities {
	v := validator.New()
	return &Activities{
		v: v,
		planCreator: &planCreatorImpl{
			v: v,
		},
	}
}
