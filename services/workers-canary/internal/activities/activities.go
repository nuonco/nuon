package activities

import "github.com/go-playground/validator/v10"

type Activities struct {
	v *validator.Validate

	TemporalHost string `validate:"required"`
}

func New(v *validator.Validate, opts ...activitiesOption) (*Activities, error) {
	a := &Activities{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	if err := v.Struct(a); err != nil {
		return nil, err
	}

	return a, nil
}

type activitiesOption func(*Activities) error

func WithTemporalHost(host string) activitiesOption {
	return func(a *Activities) error {
		a.TemporalHost = host
		return nil
	}
}
