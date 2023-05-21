package s3

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// BucketConfig configures where the state is pushed too
type BucketConfig struct {
	Name   string `json:"bucket" validate:"required"`
	Key    string `json:"key" validate:"required"`
	Region string `json:"region" validate:"required"`
}

// Credentials allow the caller to pass in pre-built credentials, from elsewhere. This is useful, for instance when this
// is being called from a waypoint plugin, but needs to write state back to the vendor's bucket
type Credentials struct {
	// credentials to use
	AWSAccessKeyID     string `json:"access_key" validate:"required"`
	AWSSecretAccessKey string `json:"secret_key" validate:"required"`
	AWSSessionToken    string `json:"token" validate:"required"`
}

// IAMConfig exposes the ability to use assume a role and use that to authenticate
type IAMConfig struct {
	// assume role arn
	RoleARN string `json:"role_arn"`

	// assume role session
	SessionTimeout time.Duration `json:"-" validate:"required"`
	SessionName    string        `json:"session_name"`
}

type s3 struct {
	v *validator.Validate

	Bucket      *BucketConfig `validate:"required,dive"`
	IAM         *IAMConfig
	Credentials *Credentials
}

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
	// TODO(jm): figure out how to authenticate a "either or" using validate-go
	if auth.IAM == nil && auth.Credentials == nil {
		return nil, fmt.Errorf("invalid options, one of IAM or Credentials must be set")
	}

	return auth, nil
}

func WithBucketConfig(bucketCfg *BucketConfig) s3Option {
	return func(s *s3) error {
		s.Bucket = bucketCfg
		return nil
	}
}

func WithIAMConfig(iamCfg *IAMConfig) s3Option {
	return func(s *s3) error {
		s.IAM = iamCfg
		return nil
	}
}

func WithCredentials(creds *Credentials) s3Option {
	return func(s *s3) error {
		s.Credentials = creds
		return nil
	}
}
