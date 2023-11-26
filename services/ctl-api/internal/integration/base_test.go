package integration

import (
	"context"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/api"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type baseIntegrationTestSuite struct {
	suite.Suite

	v         *validator.Validate
	ctx       context.Context
	ctxCancel func()

	apiClient       nuon.Client
	intAPIClient    api.Client
	githubInstallID string
}

func (s *baseIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	ctx, ctxCancel := context.WithCancel(ctx)
	s.ctx = ctx
	s.ctxCancel = ctxCancel

	s.v = validator.New()

	apiURL := os.Getenv("INTEGRATION_API_URL")
	assert.NotEmpty(s.T(), apiURL)

	apiToken := os.Getenv("INTEGRATION_API_TOKEN")
	assert.NotEmpty(s.T(), apiToken)

	apiClient, err := nuon.New(s.v,
		nuon.WithAuthToken(apiToken),
		nuon.WithURL(apiURL),
	)
	assert.NoError(s.T(), err)
	s.apiClient = apiClient

	internalAPIURL := os.Getenv("INTEGRATION_INTERNAL_API_URL")
	assert.NotEmpty(s.T(), internalAPIURL)

	intApiClient, err := api.New(s.v,
		api.WithURL(internalAPIURL),
	)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), intApiClient)
	s.intAPIClient = intApiClient

	githubInstallID := os.Getenv("INTEGRATION_GITHUB_INSTALL_ID")
	s.githubInstallID = githubInstallID
}

func (s *baseIntegrationTestSuite) createOrg() *models.AppOrg {
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)

	s.apiClient.SetOrgID(org.ID)
	if s.githubInstallID != "" {
		vcs, err := s.apiClient.CreateVCSConnection(s.ctx, &models.ServiceCreateConnectionRequest{
			GithubInstallID: generics.ToPtr(s.githubInstallID),
		})
		require.Nil(s.T(), err)
		require.NotNil(s.T(), vcs)
	}

	return org
}

func (s *baseIntegrationTestSuite) createApp(orgID string) *models.AppApp {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)

	// create app sandbox config
	cfgReq := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
	cfgReq.SandboxReleaseID = ""
	cfgReq.ConnectedGithubVcsConfig = nil

	cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, app.ID, cfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	return app
}

func (s *baseIntegrationTestSuite) createInstall(appID string) *models.AppInstall {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	install, err := s.apiClient.CreateInstall(s.ctx, appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), install)

	return install
}

func (s *baseIntegrationTestSuite) deleteOrg(orgID string) {
	disabled := os.Getenv("INTEGRATION_NO_CLEANUP")
	if disabled != "" {
		return
	}

	err := s.intAPIClient.DeleteOrg(s.ctx, orgID)
	require.NoError(s.T(), err)
}
