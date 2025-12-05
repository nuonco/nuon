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

	EnvVars  map[string]string
	FileVars map[string]interface{}
	Files    []string
}

type varsOption func(*vars) error

func New(v *validator.Validate, opts ...varsOption) (*vars, error) {
	s := &vars{
		v:        v,
		EnvVars:  make(map[string]string, 0),
		FileVars: make(map[string]any, 0),
		Files:    make([]string, 0),
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

func WithEnvVars(envVars map[string]string) varsOption {
	return func(v *vars) error {
		for k, va := range envVars {
			v.EnvVars[k] = va
		}
		return nil
	}
}

func WithFileVars(fileVars map[string]interface{}) varsOption {
	return func(v *vars) error {
		for k, va := range fileVars {
			v.FileVars[k] = va
		}
		return nil
	}
}

func WithFiles(files []string) varsOption {
	return func(v *vars) error {
		for _, f := range files {
			v.Files = append(v.Files, f)
		}
		return nil
	}
}
