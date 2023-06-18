package catalog

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

// Catalog exposes a way to get different plugin versions, based on the plugin type
type Catalog interface {
	GetLatest(context.Context, PluginType) (*Plugin, error)
	GetAll(context.Context, PluginType) ([]*Plugin, error)
}

var _ Catalog = (*catalog)(nil)

type catalog struct {
	// internal state
	v *validator.Validate

	Credentials *credentials.Config
	DevOverride bool
}

type catalogOption func(*catalog) error

func New(v *validator.Validate, opts ...catalogOption) (*catalog, error) {
	t := &catalog{v: v}
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	if err := t.v.Struct(t); err != nil {
		return nil, err
	}

	return t, nil
}

func WithCredentials(cfg *credentials.Config) catalogOption {
	return func(e *catalog) error {
		e.Credentials = cfg
		return nil
	}
}

func WithDevOverride(override bool) catalogOption {
	return func(e *catalog) error {
		e.DevOverride = override
		return nil
	}
}
