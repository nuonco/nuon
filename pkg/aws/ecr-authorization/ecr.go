package ecr

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/credentials"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock..go -source=ecr.go -package=ecr
type Client interface {
	GetAuthorization(context.Context) (*Authorization, error)
}

var _ Client = (*ecrAuthorizer)(nil)

type ecrAuthorizer struct {
	v *validator.Validate `validate:"required"`

	UseDefault bool

	Credentials *credentials.Config `validate:"-"`
}

type Option func(*ecrAuthorizer) error

func New(v *validator.Validate, opts ...Option) (*ecrAuthorizer, error) {
	auth := &ecrAuthorizer{
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

func WithRegistryID(registryID string) Option {
	return func(ecr *ecrAuthorizer) error {
		return nil
	}
}

func WithImageURL(url string) Option {
	return func(ecr *ecrAuthorizer) error {
		return nil
	}
}

func WithUseDefault(useDefault bool) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.UseDefault = useDefault
		return nil
	}
}

func WithCredentials(creds *credentials.Config) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.Credentials = creds
		return nil
	}
}

func WithRepository(repository string) Option {
	return func(ecr *ecrAuthorizer) error {
		return nil
	}
}
