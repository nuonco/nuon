package iam

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	defaultRoleSessionDuration time.Duration = time.Hour
)

type Settings struct {
	RoleARN             string `validate:"required"`
	RoleSessionName     string `validate:"required"`
	RoleSessionDuration time.Duration
}

func (s Settings) Validate(v *validator.Validate) error {
	return v.Struct(s)
}

type assumer struct {
	RoleARN             string `validate:"required"`
	RoleSessionName     string `validate:"required"`
	RoleSessionDuration time.Duration

	// internal state
	v *validator.Validate
}

type assumerOptions func(*assumer) error

// New creates a new, validated assumer with the given options
func New(v *validator.Validate, opts ...assumerOptions) (*assumer, error) {
	a := &assumer{
		RoleSessionDuration: defaultRoleSessionDuration,
	}

	if v == nil {
		return nil, fmt.Errorf("error instantiating assumer: validator is nil")
	}
	a.v = v

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	if err := a.v.Struct(a); err != nil {
		return nil, err
	}
	return a, nil
}

// WithSettings sets settings to use this to assume roles
func WithSettings(s Settings) assumerOptions {
	return func(a *assumer) error {
		if err := s.Validate(a.v); err != nil {
			return fmt.Errorf("settings are invalid: %w", err)
		}

		a.RoleARN = s.RoleARN
		a.RoleSessionName = s.RoleSessionName

		if s.RoleSessionDuration > 0 {
			a.RoleSessionDuration = s.RoleSessionDuration
		}

		return nil
	}
}

// WithRoleARN sets the ARN of the role to assume
func WithRoleARN(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleARN = s
		return nil
	}
}

// WithRoleSessionName specifies the session name to use when assuming the role
func WithRoleSessionName(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleSessionName = s
		return nil
	}
}
