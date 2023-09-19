package activities

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/workers-canary/internal"
)

type Activities struct {
	v   *validator.Validate
	cfg *internal.Config
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

func WithConfig(cfg *internal.Config) activitiesOption {
	return func(a *Activities) error {
		a.cfg = cfg
		return nil
	}
}
