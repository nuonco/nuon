package s3

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

// Package s3 exposes an archive from a tarball, stored on s3.
var _ archive.Archive = (*s3)(nil)

type s3 struct {
	v *validator.Validate

	BucketName string `validate:"required"`
	Key        string `validate:"required"`

	Credentials *credentials.Config
}

type s3Option func(*s3) error

func New(v *validator.Validate, opts ...s3Option) (*s3, error) {
	s := &s3{
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

// WithBucket name sets the s3 bucket
func WithBucketName(bucketName string) s3Option {
	return func(s *s3) error {
		s.BucketName = bucketName
		return nil
	}
}

// WithBucketKey sets the bucket key
func WithBucketKey(bucketKey string) s3Option {
	return func(s *s3) error {
		s.Key = bucketKey
		return nil
	}
}

// WithCredentials sets the credentials config
func WithCredentials(cfg *credentials.Config) s3Option {
	return func(s *s3) error {
		s.Credentials = cfg
		return nil
	}
}
