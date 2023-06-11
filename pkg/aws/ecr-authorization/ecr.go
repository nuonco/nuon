package ecr

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock..go -source=ecr.go -package=ecr
type Client interface {
	GetAuthorization(context.Context) (*Authorization, error)
}

var _ Client = (*ecrAuthorizer)(nil)

// ecrAuthorizer is a type that lets you get a token from ECR
type ecrAuthorizer struct {
	v *validator.Validate `validate:"required"`

	RegistryID string `validate:"required"`

	Credentials *credentials.Config `validate:"required"`
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

// WithRegistryID is used to set the registry id
func WithRegistryID(registryID string) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.RegistryID = registryID
		return nil
	}
}

// WithImageURL is used to determine the regsitry id
func WithImageURL(url string) Option {
	return func(ecr *ecrAuthorizer) error {
		registryID, err := parseImageURL(url)
		if err != nil {
			return err
		}

		ecr.RegistryID = registryID
		return nil
	}
}

// WithCredentials is used to set the credentials that will be used by this
func WithCredentials(creds *credentials.Config) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.Credentials = creds
		return nil
	}
}

// WithRepository is used to set the registry id by parsing the repsoitry url
func WithRepository(repository string) Option {
	return func(ecr *ecrAuthorizer) error {
		registryID, err := parseImageURL(repository)
		if err != nil {
			return fmt.Errorf("unable to parse registry id from repository: %w", err)
		}

		ecr.RegistryID = registryID
		return nil
	}
}
