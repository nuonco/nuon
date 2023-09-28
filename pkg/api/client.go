package api

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	defaultTimeout time.Duration = time.Second * 2
)

type Client interface {
	DeleteOrg(ctx context.Context, orgID string) error
}

var _ Client = (*client)(nil)

type client struct {
	v *validator.Validate

	APIURL  string `validate:"required"`
	Timeout time.Duration
}

type clientOption func(*client) error

func New(v *validator.Validate, opts ...clientOption) (*client, error) {
	c := &client{v: v,
		Timeout: defaultTimeout,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if err := c.v.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

// WithURL specifies the url to use
func WithURL(url string) clientOption {
	return func(c *client) error {
		c.APIURL = url
		return nil
	}
}

// WithTimeout specifies the timeout to use
func WithTimeout(dur time.Duration) clientOption {
	return func(c *client) error {
		c.Timeout = dur
		return nil
	}
}
