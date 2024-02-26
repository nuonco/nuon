package json

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

// Package json exposes an archive from a tarball, stored on json.
var _ archive.Archive = (*json)(nil)

type json struct {
	v *validator.Validate

	FileName string `validate:"required"`
	Byts     []byte `validate:"required"`
}

type jsonOption func(*json) error

func New(v *validator.Validate, opts ...jsonOption) (*json, error) {
	s := &json{
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

// WithFileame name sets the json filename
func WithFileName(fileName string) jsonOption {
	return func(d *json) error {
		d.FileName = fileName
		return nil
	}
}

// WithFileame name sets the json filename
func WithJSON(byts []byte) jsonOption {
	return func(d *json) error {
		d.Byts = byts
		return nil
	}
}
