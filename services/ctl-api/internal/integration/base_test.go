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

	// setup internal api
	internalAPIURL := os.Getenv("INTEGRATION_INTERNAL_API_URL")
	assert.NotEmpty(s.T(), internalAPIURL)

	intApiClient, err := api.New(s.v,
		api.WithURL(internalAPIURL),
	)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), intApiClient)
	s.intAPIClient = intApiClient

	// create integration user
	intUser, err := s.intAPIClient.CreateIntegrationUser(ctx)
	require.NoError(s.T(), err)

	apiURL := os.Getenv("INTEGRATION_API_URL")
	assert.NotEmpty(s.T(), apiURL)

	apiClient, err := nuon.New(s.v,
		nuon.WithAuthToken(intUser.APIToken),
		nuon.WithURL(apiURL),
	)
	assert.NoError(s.T(), err)
	s.apiClient = apiClient

	s.githubInstallID = intUser.GithubInstallID
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

func (s *baseIntegrationTestSuite) createAppInputConfig(appID string) *models.AppAppInputConfig {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
	appInputConfig, err := s.apiClient.CreateAppInputConfig(s.ctx, appID, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), appInputConfig)

	return appInputConfig
}

func (s *baseIntegrationTestSuite) fakeInstallInputsForApp(appID string) map[string]string {
	inputCfg, err := s.apiClient.GetAppInputLatestConfig(s.ctx, appID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), inputCfg)

	vals := make(map[string]string, 0)
	for _, input := range inputCfg.AppInputs {
		vals[input.Name] = generics.GetFakeObj[string]()
	}

	return vals
}

func (s *baseIntegrationTestSuite) createAppWithInputs(orgID string) *models.AppApp {
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

	inputReq := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
	_, err = s.apiClient.CreateAppInputConfig(s.ctx, app.ID, inputReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	return app
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
