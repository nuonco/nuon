package auth

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
)

// Package vars exposes an archive that loads a terraform archive from an vars artifact
var _ variables.Variables = (*auth)(nil)

type auth struct {
	v *validator.Validate

	Auth *credentials.Config `validate:"required"`
}

type varsOption func(*auth) error

func New(v *validator.Validate, opts ...varsOption) (*auth, error) {
	s := &auth{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("unable to set %d option: %w", idx, err)
		}
	}
	if err := s.v.Struct(s); err != nil {
		return nil, err
	}

	return s, nil
}

func WithAuth(cfg *credentials.Config) varsOption {
	return func(v *auth) error {
		v.Auth = cfg
		return nil
	}
}
