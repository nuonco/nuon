package plan

import "github.com/go-playground/validator/v10"

type Activities struct {
	planCreator planCreator
}

func NewActivities() *Activities {
	v := validator.New()
	return &Activities{
		planCreator: &planCreatorImpl{
			v: v,
		},
	}
}
