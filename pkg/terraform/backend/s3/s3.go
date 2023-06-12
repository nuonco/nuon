package s3

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
)

// BucketConfig configures where the state is pushed too
type BucketConfig struct {
	Name   string `mapstructure:"bucket" validate:"required"`
	Key    string `mapstructure:"key" validate:"required"`
	Region string `mapstructure:"region" validate:"required"`
}

type s3 struct {
	v *validator.Validate

	Bucket      *BucketConfig       `validate:"required,dive"`
	Credentials *credentials.Config `validate:"-"`
}

var _ backend.Backend = (*s3)(nil)

type s3Option func(*s3) error

func New(v *validator.Validate, opts ...s3Option) (*s3, error) {
	auth := &s3{
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

func WithBucketConfig(bucketCfg *BucketConfig) s3Option {
	return func(s *s3) error {
		s.Bucket = bucketCfg
		return nil
	}
}

func WithCredentials(creds *credentials.Config) s3Option {
	return func(s *s3) error {
		if err := creds.Validate(s.v); err != nil {
			return fmt.Errorf("unable to validate credentials: %w", err)
		}

		s.Credentials = creds
		return nil
	}
}
