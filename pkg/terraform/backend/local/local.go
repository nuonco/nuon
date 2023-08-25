package local

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
)

type local struct {
	v *validator.Validate

	Fp string `validate:"required"`
}

var _ backend.Backend = (*local)(nil)

type localOption func(*local) error

func New(v *validator.Validate, opts ...localOption) (*local, error) {
	auth := &local{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(auth); err != nil {
			return nil, err
		}
	}

	if err := auth.v.Struct(auth); err != nil {
		return nil, err
	}

	return auth, nil
}

func WithFilepath(fp string) localOption {
	return func(l *local) error {
		l.Fp = fp
		return nil
	}
}
