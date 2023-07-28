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
	GetOrg(ctx context.Context, orgID string) (*Org, error)
	GetOrgs(ctx context.Context, orgID string) ([]*Org, error)
	UpsertOrg(ctx context.Context, input OrgInput) (*Org, error)
	DeleteOrg(ctx context.Context, orgID string) error

	// apps
	GetApp(ctx context.Context, appID string) (*App, error)
	GetApps(ctx context.Context, orgID string) ([]*App, error)
	UpsertApp(ctx context.Context, input AppInput) (*App, error)
	DeleteApp(ctx context.Context, appID string) (bool, error)

	// components
	GetComponent(ctx context.Context, componentID string) (*Component, error)
	UpsertComponent(ctx context.Context, input ComponentInput) (*Component, error)
	DeleteComponent(ctx context.Context, id string) (bool, error)
	GetComponents(ctx context.Context, appID string) ([]*Component, error)

	// builds
	GetBuild(ctx context.Context, installID string) (*Build, error)
	GetBuilds(ctx context.Context, componentID string) ([]*Build, error)
	StartBuild(ctx context.Context, input BuildInput) (*Build, error)
	CancelBuild(ctx context.Context, installID string) (bool, error)
	GetBuildStatus(ctx context.Context, buildID string) (Status, error)

	// installs
	GetInstall(ctx context.Context, installID string) (*Install, error)
	GetInstalls(ctx context.Context, appID string) ([]*Install, error)
	UpsertInstall(ctx context.Context, input InstallInput) (*Install, error)
	DeleteInstall(ctx context.Context, installID string) (bool, error)
	GetInstallStatus(ctx context.Context, orgID, appID, installID string) (Status, error)

	// deploys
	StartDeploy(ctx context.Context, input DeployInput) (*Deploy, error)
	GetDeploy(ctx context.Context, deployID string) (*Deploy, error)

	// instance
	GetInstanceStatus(ctx context.Context, installID, componentID, deployID string) (Status, error)

	// users
	GetCurrentUser(ctx context.Context) (*getCurrentUserMeUser, error)
	GetConnectedRepos(ctx context.Context, orgID string) ([]*ConnectedRepo, error)
	GetConnectedRepo(ctx context.Context, orgID string, repoName string) (*ConnectedRepo, error)
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
