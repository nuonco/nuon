package gqlclient

import (
	"context"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/go-playground/validator/v10"
)

//go:generate -command genqlient go run github.com/Khan/genqlient
//go:generate genqlient genqlient.yaml

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock.go -source=client.go -package=gqlclient
type Client interface {
	// orgs
	GetOrg(ctx context.Context, orgID string) (*getOrgOrg, error)
	GetOrgs(ctx context.Context, orgID string) ([]*getOrgsOrgsOrgConnectionEdgesOrgEdgeNodeOrg, error)
	UpsertOrg(ctx context.Context, input OrgInput) (*upsertOrgUpsertOrg, error)
	DeleteOrg(ctx context.Context, orgID string) error

	// apps
	GetApp(ctx context.Context, appID string) (*getAppApp, error)
	GetApps(ctx context.Context, orgID string) ([]*getAppsAppsAppConnectionEdgesAppEdgeNodeApp, error)
	UpsertApp(ctx context.Context, input AppInput) (*upsertAppUpsertApp, error)
	DeleteApp(ctx context.Context, appID string) error

	// installs
	GetInstall(ctx context.Context, installID string) (*getInstallInstall, error)
	GetInstalls(ctx context.Context, appID string) ([]*getInstallsInstallsInstallConnectionEdgesInstallEdgeNodeInstall, error)

	// components
	GetComponent(ctx context.Context, componentID string) (*getComponentComponent, error)
	GetComponents(ctx context.Context, appID string) ([]*getComponentsComponentsComponentConnectionEdgesComponentEdgeNodeComponent, error)

	// users
	GetCurrentUser(ctx context.Context) (*getCurrentUserMeUser, error)
}

var _ Client = (*client)(nil)

type client struct {
	v *validator.Validate

	AuthToken string `validate:"required"`
	URL       string `validate:"required"`

	// inner fields
	httpClient    *http.Client
	graphqlClient graphql.Client
}

type clientOption func(*client) error

func New(v *validator.Validate, opts ...clientOption) (*client, error) {
	c := &client{v: v}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.httpClient = &http.Client{
		Transport: &authTransport{
			authToken: c.AuthToken,
			transport: http.DefaultTransport,
		},
	}
	c.graphqlClient = graphql.NewClient(c.URL, c.httpClient)

	if err := c.v.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

// WithAuthToken specifies the auth token to use
func WithAuthToken(token string) clientOption {
	return func(c *client) error {
		c.AuthToken = token
		return nil
	}
}

// WithURL specifies the url to use
func WithURL(url string) clientOption {
	return func(c *client) error {
		c.URL = url
		return nil
	}
}
