package mailchimp

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Client interface {
	UpsertListMember(ctx context.Context, email string) error
}

type client struct {
	v *validator.Validate

	APIKey       string `validate:"required"`
	ServerPrefix string `validate:"required"`
	AudienceID   string `validate:"required"`

	baseURL string
}

var _ Client = (*client)(nil)

func New(v *validator.Validate, opts ...clientOption) (*client, error) {
	c := &client{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := c.v.Struct(c); err != nil {
		return nil, fmt.Errorf("unable to validate client: %w", err)
	}

	if c.baseURL == "" {
		c.baseURL = fmt.Sprintf("https://%s.api.mailchimp.com/3.0", c.ServerPrefix)
	}

	return c, nil
}

type clientOption func(*client) error

func WithAPIKey(apiKey string) clientOption {
	return func(c *client) error {
		c.APIKey = apiKey
		return nil
	}
}

func WithServerPrefix(serverPrefix string) clientOption {
	return func(c *client) error {
		c.ServerPrefix = serverPrefix
		return nil
	}
}

func WithAudienceID(audienceID string) clientOption {
	return func(c *client) error {
		c.AudienceID = audienceID
		return nil
	}
}

// WithBaseURL overrides the derived Mailchimp API base URL, for tests.
func WithBaseURL(baseURL string) clientOption {
	return func(c *client) error {
		c.baseURL = baseURL
		return nil
	}
}
