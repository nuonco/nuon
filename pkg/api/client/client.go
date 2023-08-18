package gqlclient

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

//go:generate -command swagger go run github.com/go-swagger/go-swagger/cmd/swagger
//go:generate swagger generate client --skip-tag-packages -f ../../../services/ctl-api/docs/swagger.json
type Client interface {
	// orgs
	GetOrg(ctx context.Context, orgID string) (*models.AppOrg, error)
	GetOrgs(ctx context.Context) ([]*models.AppOrg, error)
	CreateOrg(ctx context.Context, req *models.ServiceCreateOrgRequest) ([]*models.AppOrg, error)
	UpdateOrg(ctx context.Context, req *models.ServiceUpdateOrgRequest) ([]*models.AppOrg, error)
	CreateOrgUser(ctx context.Context, req *models.ServiceCreateOrgUserRequest) ([]*models.AppOrg, error)

	// internal methods
	GetApp(ctx context.Context, appID string) (*models.AppApp, error)
	GetApps(ctx context.Context, orgID string) ([]*models.AppApp, error)
	CreateApp(ctx context.Context, req *models.ServiceCreateAppRequest) (*models.AppApp, error)
	UpdateApp(ctx context.Context, req *models.ServiceUpdateAppRequest) (*models.AppApp, error)
	DeleteApp(ctx context.Context, appID string) (bool, error)
	UpdateAppSandbox(ctx context.Context, appID string, req *models.ServiceUpdateAppSandboxRequest) (bool, error)

	// general methods
	GetCurrentUser(ctx context.Context) (*models.AppUserToken, error)
	PublishMetrics(ctx context.Context, req []*models.ServicePublishMetricInput) (*models.AppUserToken, error)

	// sandbox methods
	GetSandboxes(ctx context.Context, sandboxID string) ([]*models.AppSandbox, error)
	GetSandbox(ctx context.Context, sandboxID string) (*models.AppSandbox, error)
	GetSandboxReleases(ctx context.Context, sandboxID string) (*models.AppSandboxRelease, error)

	// vcs connections
	CreateOrgVCSConnection(ctx context.Context, orgID string, req *models.ServiceCreateOrgConnectionRequest) (*models.AppVCSConnection, error)
	GetOrgVCSConnectedRepos(ctx context.Context, orgID string) ([]*models.ServiceRepository, error)

	// installs
	CreateInstall(ctx context.Context, appID string, req *models.ServiceCreateInstallRequest) (models.AppInstall, error)
	GetAppInstalls(ctx context.Context, appID string) ([]*models.AppInstall, error)
	GetAllInstalls(ctx context.Context) ([]*models.AppInstall, error)

	GetInstall(ctx context.Context, installID string) ([]*models.AppInstall, error)
	UpdateInstall(ctx context.Context, installID string, req *models.ServiceUpdateInstallRequest) (*models.AppInstall, error)
	DeleteInstall(ctx context.Context, installID string) (bool, error)

	// install deploys
	GetInstallDeploys(ctx context.Context, installID string) ([]*models.AppInstallDeploy, error)
	CreateInstallDeploy(ctx context.Context, installID string, req *models.AppInstallDeploy) (*models.AppInstallDeploy, error)
	GetInstallDeploy(ctx context.Context, installID, deployID string) (*models.AppInstallDeploy, error)
	GetInstallDeployLogs(ctx context.Context, installID, deployID string) ([]*models.ServiceDeployLog, error)

	// install components
	GetInstallComponents(ctx context.Context, installID string) ([]*models.AppInstallComponent, error)
	GetInstallComponentDeploys(ctx context.Context, installID, componentID string) ([]*models.AppInstallDeploy, error)

	// components
	GetAllComponent(ctx context.Context) ([]*models.AppComponent, error)
	GetAppComponent(ctx context.Context, appID string) ([]*models.AppComponent, error)
	CreateComponent(ctx context.Context, appID string, req *models.ServiceCreateComponentRequest) (*models.AppComponent, error)

	GetComponent(ctx context.Context, componentID string) (*models.AppComponent, error)
	UpdateComponent(ctx context.Context, componentID string, req *models.ServiceCreateComponentRequest) (*models.AppComponent, error)
	DeleteComponent(ctx context.Context, componentID string) (bool, error)

	// component configs
	CreateTerraformModuleComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateTerraformModuleComponentConfigRequest) (*models.AppComponentConfigConnection, error)
	CreateHelmComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateHelmComponentConfigRequest) (*models.AppComponentConfigConnection, error)
	CreateDockerBuildComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateDockerBuildComponentConfigRequest) (*models.AppComponentConfigConnection, error)
	CreateExternalImageComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateExternalImageComponentConfigRequest) (*models.AppComponentConfigConnection, error)
	GetComponentConfigs(ctx context.Context, componentID string) ([]*models.AppComponentConfigConnection, error)
	GetComponentLatestConfig(ctx context.Context, componentID string) (*models.AppComponentConfigConnection, error)

	// builds
	CreateComponentBuild(ctx context.Context, componentID string, req *models.ServiceCreateComponentBuildRequest) (*models.AppComponentBuild, error)
	GetComponentBuilds(ctx context.Context, componentID string) ([]*models.AppComponentBuild, error)
	GetComponentLatestBuild(ctx context.Context, componentID string) (*models.AppComponentBuild, error)
	GetComponentBuild(ctx context.Context, componentID, buildID string) (*models.AppComponentBuild, error)
	GetComponentBuildLogs(ctx context.Context, componentID, buildID string) ([]*models.ServiceBuildLog, error)
}

//var _ Client = (*client)(nil)

type client struct {
	v *validator.Validate

	APIToken string `validate:"required"`
	URL      string `validate:"required"`
	OrgID    string

	// inner fields
	httpClient *http.Client
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
			authToken: c.APIToken,
			transport: http.DefaultTransport,
		},
	}

	if err := c.v.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

// WithAuthToken specifies the auth token to use
func WithAuthToken(token string) clientOption {
	return func(c *client) error {
		c.APIToken = token
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

// WithOrgID specifies the org id to use
func WithOrgID(orgID string) clientOption {
	return func(c *client) error {
		c.OrgID = orgID
		return nil
	}
}
