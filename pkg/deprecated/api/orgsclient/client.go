package orgsclient

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/types/orgs-api/apps/v1/appsv1connect"
	"github.com/powertoolsdev/mono/pkg/types/orgs-api/builds/v1/buildsv1connect"
	"github.com/powertoolsdev/mono/pkg/types/orgs-api/installs/v1/installsv1connect"
	"github.com/powertoolsdev/mono/pkg/types/orgs-api/instances/v1/instancesv1connect"
	"github.com/powertoolsdev/mono/pkg/types/orgs-api/orgs/v1/orgsv1connect"
)

type Client struct {
	Apps      appsv1connect.AppsServiceClient           `validate:"required"`
	Builds    buildsv1connect.BuildsServiceClient       `validate:"required"`
	Installs  installsv1connect.InstallsServiceClient   `validate:"required"`
	Instances instancesv1connect.InstancesServiceClient `validate:"required"`
	Orgs      orgsv1connect.OrgsServiceClient           `validate:"required"`

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
	a.Apps = appsv1connect.NewAppsServiceClient(http.DefaultClient, a.Addr)
	a.Builds = buildsv1connect.NewBuildsServiceClient(http.DefaultClient, a.Addr)
	a.Installs = installsv1connect.NewInstallsServiceClient(http.DefaultClient, a.Addr)
	a.Instances = instancesv1connect.NewInstancesServiceClient(http.DefaultClient, a.Addr)
	a.Orgs = orgsv1connect.NewOrgsServiceClient(http.DefaultClient, a.Addr)
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
