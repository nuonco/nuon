package loops

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Client interface {
	SendTransactionalEmail(ctx context.Context, email, transactionalEmailID string, vars map[string]interface{}) error
}

type client struct {
	v *validator.Validate

	APIKey string `validate:"required"`
}

var _ Client = (*client)(nil)

// New returns a default client, which emits metrics to statsd by default
func New(v *validator.Validate, opts ...clientOption) (*client, error) {

	r := &client{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate client: %w", err)
	}

	return r, nil
}

type clientOption func(*client) error

func WithAPIKey(apiKey string) clientOption {
	return func(w *client) error {
		w.APIKey = apiKey
		return nil
	}
}
