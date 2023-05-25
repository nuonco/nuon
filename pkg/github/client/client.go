package client

import (
	"fmt"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
)

type Option func(*client) error

type client struct {
	v *validator.Validate

	HTTPClient *http.Client `validate:"required"`
	AppID      int64        `validate:"required"`
	AppKey     []byte       `validate:"required"`
}

func New(v *validator.Validate, opts ...Option) (*github.Client, error) {
	gh := &client{
		v:          v,
		HTTPClient: &http.Client{},
	}

	for idx, opt := range opts {
		if err := opt(gh); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := gh.v.Struct(gh); err != nil {
		return nil, fmt.Errorf("unable to validate github: %w", err)
	}

	appstp, err := ghinstallation.NewAppsTransport(http.DefaultTransport, gh.AppID, gh.AppKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create github apps transport: %w", err)
	}

	gh.HTTPClient.Transport = appstp
	return github.NewClient(gh.HTTPClient), nil
}

func WithAppID(appID string) Option {
	return func(gh *client) error {
		githubAppID, err := strconv.ParseInt(appID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid github app id: %w", err)
		}
		gh.AppID = githubAppID
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(gh *client) error {
		gh.HTTPClient = httpClient
		return nil
	}
}

func WithAppKey(key []byte) Option {
	return func(gh *client) error {
		gh.AppKey = key
		return nil
	}
}
