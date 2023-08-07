package health

import "github.com/go-playground/validator/v10"

type svc struct {
	v *validator.Validate
}

func New(v *validator.Validate) (*svc, error) {
	return &svc{
		v: v,
	}, nil
}
