package mailchimp

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	mcclient "github.com/mailchimp/mailchimp-marketing-go-sdk/client"
	"github.com/mailchimp/mailchimp-marketing-go-sdk/option"
)

type Client interface {
	Enabled() bool
	UpsertListMember(ctx context.Context, email string) error
}

type client struct {
	v *validator.Validate

	APIKey       string `validate:"required"`
	AudienceID   string `validate:"required"`
	ServerPrefix string

	baseURL string
	sdk     *mcclient.MailchimpClient
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

	if c.baseURL == "" && c.ServerPrefix != "" {
		c.baseURL = fmt.Sprintf("https://%s.api.mailchimp.com", c.ServerPrefix)
	}

	sdkOpts := []option.RequestOption{
		option.WithToken(c.APIKey),
		option.WithMaxAttempts(1),
	}
	if c.baseURL != "" {
		sdkOpts = append(sdkOpts, option.WithBaseURL(c.baseURL))
	}
	c.sdk = mcclient.NewMailchimpClient(sdkOpts...)

	return c, nil
}

func (c *client) Enabled() bool {
	return true
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

func WithBaseURL(baseURL string) clientOption {
	return func(c *client) error {
		c.baseURL = baseURL
		return nil
	}
}
