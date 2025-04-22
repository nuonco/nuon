package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
)

type HTTPBackendConfig struct {
	Address       string `json:"address"`
	LockAddress   string `json:"lock_address"`
	UnlockAddress string `json:"unlock_address"`
	LockMethod    string `json:"lock_method"`
	UnlockMethod  string `json:"unlock_method"`
}

type http struct {
	v *validator.Validate

	Config *HTTPBackendConfig `validate:"required"`
}

var _ backend.Backend = (*http)(nil)

type httpOption func(*http) error

func New(v *validator.Validate, opts ...httpOption) (*http, error) {
	auth := &http{
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

func WithHTTPConfig(cfg *HTTPBackendConfig) httpOption {
	return func(s *http) error {
		s.Config = cfg
		return nil
	}
}
