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
	ListOrgs(ctx context.Context) ([]Org, error)
	DeleteOrg(ctx context.Context, orgID string) error
	DeprovisionOrg(ctx context.Context, orgID string) error
	ReprovisionOrg(ctx context.Context, orgID string) error
	RestartOrg(ctx context.Context, orgID string) error
	AddSupportUsers(ctx context.Context, orgID string) error

	ListApps(ctx context.Context) ([]App, error)
	ReprovisionApp(ctx context.Context, appID string) error
	RestartApp(ctx context.Context, appID string) error
	UpdateAppSandbox(ctx context.Context, appID string) error

	ListInstalls(ctx context.Context) ([]Install, error)
	ReprovisionInstall(ctx context.Context, installID string) error
	RestartInstall(ctx context.Context, installID string) error
	DeprovisionInstall(ctx context.Context, installID string) error
	DeleteInstall(ctx context.Context, installID string) error
	UpdateInstallSandbox(ctx context.Context, installID string) error

	ListComponents(ctx context.Context) ([]Component, error)
	RestartComponent(ctx context.Context, componentID string) error

	ListReleases(ctx context.Context) ([]Release, error)
	RestartRelease(ctx context.Context, releaseID string) error

	ProvisionCanary(ctx context.Context, sandboxMode bool) error
	DeprovisionCanary(ctx context.Context, canaryID string) error
	StartCanaryCron(ctx context.Context) error
	StopCanaryCron(ctx context.Context) error

	CreateIntegrationUser(ctx context.Context) (*CreateIntegrationUserResponse, error)
	CreateCanaryUser(ctx context.Context, canaryID string) (*CreateCanaryUserResponse, error)
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
