package dir

import (
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/terraform/archive"
)

var _ archive.Archive = (*dir)(nil)

type dir struct {
	v *validator.Validate

	Path                     string `validate:"required"`
	IgnoreTerraformLockFile  bool
	IgnoreTerraformStateFile bool
	IgnoreDotTerraformDir    bool

	AddBackendFile bool
	AddBackendType string
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

func WithPath(path string) dirOption {
	return func(d *dir) error {
		path, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("unable to resolve path to absolute: %w", err)
		}

		d.Path = path

		return nil
	}
}

func WithIgnoreTerraformLockFile() dirOption {
	return func(d *dir) error {
		d.IgnoreTerraformLockFile = true
		return nil
	}
}

func WithIgnoreDotTerraformDir() dirOption {
	return func(d *dir) error {
		d.IgnoreDotTerraformDir = true
		return nil
	}
}

func WithIgnoreTerraformStateFile() dirOption {
	return func(d *dir) error {
		d.IgnoreTerraformStateFile = true
		return nil
	}
}

func WithAddBackendFile(typ string) dirOption {
	return func(d *dir) error {
		d.AddBackendFile = true
		d.AddBackendType = typ
		return nil
	}
}
