package static

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
)

// Package vars exposes an archive that loads a terraform archive from an vars artifact
var _ variables.Variables = (*vars)(nil)

type vars struct {
	v *validator.Validate
}

type varsOption func(*vars) error

func New(v *validator.Validate, opts ...varsOption) (*vars, error) {
	s := &vars{
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
