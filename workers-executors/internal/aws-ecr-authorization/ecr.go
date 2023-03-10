package ecr

import (
	"github.com/go-playground/validator/v10"
)

// ecrAuthorizer is a type that lets you get a token from ECR
type ecrAuthorizer struct {
	v *validator.Validate `validate:"required"`

	RegistryID            string `validate:"required"`
	AssumeRoleArn         string `validate:"required"`
	AssumeRoleSessionName string `validate:"required"`
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

// WithAssumeRoleArn is used to set the iam role arn to auth with
func WithAssumeRoleArn(roleArn string) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.AssumeRoleArn = roleArn
		return nil
	}
}

// WithAssumeRoleSessionName is used to set the session name
func WithAssumeRoleSessionName(sessionName string) Option {
	return func(ecr *ecrAuthorizer) error {
		ecr.AssumeRoleSessionName = sessionName
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
