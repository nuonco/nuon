package eksclient

import (
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

type eksClient struct {
	RoleARN         string
	RoleSessionName string
	AWSAuth         *credentials.Config

	ClusterName string `validate:"required"`
	Region      string `validate:"required"`

	// internal state
	v *validator.Validate
}

type eksOptions func(*eksClient) error

// New creates a new, validated eks with the given options
func New(v *validator.Validate, opts ...eksOptions) (*eksClient, error) {
	e := &eksClient{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}
	if err := e.v.Struct(e); err != nil {
		return nil, err
	}
	return e, nil
}

// WithCredentials
func WithCredentials(cfg *credentials.Config) eksOptions {
	return func(e *eksClient) error {
		e.AWSAuth = cfg
		e.Region = cfg.Region
		return nil
	}
}

// WithRoleARN sets the ARN of the role to assume
func WithRoleARN(s string) eksOptions {
	return func(e *eksClient) error {
		e.RoleARN = s
		return nil
	}
}

// WithRoleSessionName specifies the session name to use when assuming the role
func WithRoleSessionName(s string) eksOptions {
	return func(e *eksClient) error {
		e.RoleSessionName = s
		return nil
	}
}

// WithClusterName specifies the session name to use when assuming the role
func WithClusterName(s string) eksOptions {
	return func(e *eksClient) error {
		e.ClusterName = s
		return nil
	}
}

// WithRegion specifies the session name to use when assuming the role
func WithRegion(s string) eksOptions {
	return func(e *eksClient) error {
		e.Region = s
		return nil
	}
}
