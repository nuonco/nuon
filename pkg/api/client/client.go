package client

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/types/api/admin/v1/adminv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/app/v1/appv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/build/v1/buildv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/component/v1/componentv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/deploy/v1/deployv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/deployment/v1/deploymentv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/github/v1/githubv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/install/v1/installv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/org/v1/orgv1connect"
	"github.com/powertoolsdev/mono/pkg/types/api/user/v1/userv1connect"
)

type Client struct {
	Admin      adminv1connect.AdminServiceClient            `validate:"required"`
	Apps       appv1connect.AppsServiceClient               `validate:"required"`
	Builds     buildv1connect.BuildsServiceClient           `validate:"required"`
	Components componentv1connect.ComponentsServiceClient   `validate:"required"`
	Deploys    deployv1connect.DeployServiceClient          `validate:"required"`
	Deployment deploymentv1connect.DeploymentsServiceClient `validate:"required"`
	Github     githubv1connect.GithubServiceClient          `validate:"required"`
	Installs   installv1connect.InstallsServiceClient       `validate:"required"`
	Orgs       orgv1connect.OrgsServiceClient               `validate:"required"`
	Users      userv1connect.UsersServiceClient             `validate:"required"`

	// internal fields for managing connection
	v          *validator.Validate `validate:"required"`
	Addr       string              `validate:"required"`
	HttpClient *http.Client        `validate:"required"`
}

type apiOption func(*Client) error

func New(v *validator.Validate, opts ...apiOption) (*Client, error) {
	a := &Client{
		v:          v,
		HttpClient: http.DefaultClient,
	}

	for idx, opt := range opts {
		if err := opt(a); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	// initialize each client
	a.Admin = adminv1connect.NewAdminServiceClient(http.DefaultClient, a.Addr)
	a.Apps = appv1connect.NewAppsServiceClient(http.DefaultClient, a.Addr)
	a.Builds = buildv1connect.NewBuildsServiceClient(http.DefaultClient, a.Addr)
	a.Components = componentv1connect.NewComponentsServiceClient(http.DefaultClient, a.Addr)
	a.Deploys = deployv1connect.NewDeployServiceClient(http.DefaultClient, a.Addr)
	a.Deployment = deploymentv1connect.NewDeploymentsServiceClient(http.DefaultClient, a.Addr)
	a.Github = githubv1connect.NewGithubServiceClient(http.DefaultClient, a.Addr)
	a.Installs = installv1connect.NewInstallsServiceClient(http.DefaultClient, a.Addr)
	a.Orgs = orgv1connect.NewOrgsServiceClient(http.DefaultClient, a.Addr)
	a.Users = userv1connect.NewUsersServiceClient(http.DefaultClient, a.Addr)

	if err := v.Struct(a); err != nil {
		return nil, fmt.Errorf("unable to validate api: %w", err)
	}
	return a, nil
}

func WithAddr(addr string) apiOption {
	return func(a *Client) error {
		a.Addr = addr
		return nil
	}
}

func WithHTTPClient(client *http.Client) apiOption {
	return func(a *Client) error {
		a.HttpClient = client
		return nil
	}
}
