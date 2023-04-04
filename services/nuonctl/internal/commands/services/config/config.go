package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ServiceType string

const (
	ServiceTypeUnknown     = ""
	ServiceTypeWorker      = "worker"
	ServiceTypeInternalApi = "internal-api"
	ServiceTypeBinary      = "binary"
)

// Config represents the configurations for a target
type Config struct {
	Env  map[string]string `yaml:"env"`
	Type ServiceType       `yaml:"type" validate:"required"`
	Port int               `yaml:"port"`
}

func (c *Config) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

type loader struct {
	v *validator.Validate

	RootDir string `validate:"required"`
	Service string `validate:"required"`
}

type loaderOption func(*loader) error

func New(v *validator.Validate, opts ...loaderOption) (*loader, error) {
	l := &loader{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(l); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := l.v.Struct(l); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	return l, nil
}

// WithRootDir sets the root directory
func WithRootDir(s string) loaderOption {
	return func(l *loader) error {
		l.RootDir = s
		return nil
	}
}

// WithService sets the service
func WithService(s string) loaderOption {
	return func(l *loader) error {
		l.Service = s
		return nil
	}
}
