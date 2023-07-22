package dir

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

// Package dir exposes an archive from a tarball, stored on dir.
var _ archive.Archive = (*dir)(nil)

type dir struct {
	v *validator.Validate

	Path string `validate:"required"`
}

type dirOption func(*dir) error

func New(v *validator.Validate, opts ...dirOption) (*dir, error) {
	s := &dir{
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

// WithPath name sets the dir path
func WithPath(path string) dirOption {
	return func(d *dir) error {
		d.Path = path
		return nil
	}
}
