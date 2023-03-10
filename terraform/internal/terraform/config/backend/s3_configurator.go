package backend

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

type S3Config struct {
	BucketName   string `json:"bucket" validate:"required"`
	BucketKey    string `json:"key" validate:"required"`
	BucketRegion string `json:"region" validate:"required"`
}

type s3Configurator struct {
	S3BackendConfig *S3Config `validate:"required,dive"`

	// internal state
	validator *validator.Validate
}

type s3ConfiguratorOption func(*s3Configurator) error

func NewS3Configurator(v *validator.Validate, opts ...s3ConfiguratorOption) (*s3Configurator, error) {
	s := &s3Configurator{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating s3 configurator: validator is nil")
	}
	s.validator = v

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if err := s.validator.Struct(s); err != nil {
		return nil, err
	}
	return s, nil
}

func WithBackendConfig(c *S3Config) s3ConfiguratorOption {
	return func(s *s3Configurator) error {
		s.S3BackendConfig = c
		return nil
	}
}

func (s *s3Configurator) JSON(w io.Writer) error {
	byts, err := json.Marshal(s.S3BackendConfig)
	if err != nil {
		return err
	}

	_, err = w.Write(byts)
	return err
}
