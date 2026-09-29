package eksclient

import (
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/credentials"
)

type eksClient struct {
	AWSAuth *credentials.Config

	ClusterName string `validate:"required"`
	Region      string `validate:"required"`

	v *validator.Validate
}

type eksOptions func(*eksClient) error

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

func WithCredentials(cfg *credentials.Config) eksOptions {
	return func(e *eksClient) error {
		e.AWSAuth = cfg
		e.Region = cfg.Region
		return nil
	}
}

func WithClusterName(s string) eksOptions {
	return func(e *eksClient) error {
		e.ClusterName = s
		return nil
	}
}
