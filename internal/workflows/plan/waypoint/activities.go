package plan

import "github.com/go-playground/validator/v10"

type Activities struct {
	v            *validator.Validate
	planCreator  planCreator
	planUploader planUploader
}

func NewActivities() *Activities {
	v := validator.New()
	return &Activities{
		v: v,
		planCreator: &planCreatorImpl{
			v: v,
		},
		planUploader: &planUploaderImpl{},
	}
}
