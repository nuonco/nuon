package plan

import "github.com/go-playground/validator/v10"

type Activities struct {
	v            *validator.Validate
	planUploader planUploader
}

func NewActivities() *Activities {
	v := validator.New()
	return &Activities{
		v:            v,
		planUploader: &planUploaderImpl{},
	}
}
