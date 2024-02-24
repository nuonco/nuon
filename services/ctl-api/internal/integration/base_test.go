package integration

import (
	"context"
	"os"
	"strings"

	"github.com/avast/retry-go"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/api"
	"github.com/powertoolsdev/mono/pkg/generics"
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
	require.NotEmpty(s.T(), internalAPIURL)

	intApiClient, err := api.New(s.v,
		api.WithURL(internalAPIURL),
	)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), intApiClient)
	s.intAPIClient = intApiClient

	// create integration user, while retrying up to 5 times due to twingate network instability inside of GHA.
	var intUser *api.CreateIntegrationUserResponse
	err = retry.Do(func() error {
		intUser, err = s.intAPIClient.CreateIntegrationUser(s.ctx)
		return err
	}, retry.Attempts(5))
	require.NoError(s.T(), err)

	apiURL := os.Getenv("INTEGRATION_API_URL")
	require.NotEmpty(s.T(), apiURL)

	apiClient, err := nuon.New(s.v,
		nuon.WithAuthToken(intUser.APIToken),
		nuon.WithURL(apiURL),
	)
	require.NoError(s.T(), err)
	s.apiClient = apiClient

	s.githubInstallID = intUser.GithubInstallID
}

func (s *baseIntegrationTestSuite) fakeOrgRequest() *models.ServiceCreateOrgRequest {
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	orgReq.UseSandboxMode = true
	return orgReq
}

func (s *baseIntegrationTestSuite) createOrg() *models.AppOrg {
	orgReq := s.fakeOrgRequest()

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
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)

	s.createAppSandboxConfig(app.ID)
	s.createAppRunnerConfig(app.ID)

	inputReq := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
	inputReq.Inputs[generics.GetFakeObj[string]()] = generics.GetFakeObj[models.ServiceAppInputRequest]()
	inputReq.Inputs = s.formatInputs(inputReq.Inputs)

	cfg, err := s.apiClient.CreateAppInputConfig(s.ctx, app.ID, inputReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	return app
}

func (s *baseIntegrationTestSuite) createAppSandboxConfig(appID string) {
	// create app sandbox config
	cfgReq := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
	cfgReq.SandboxReleaseID = ""
	cfgReq.ConnectedGithubVcsConfig = nil

	cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, appID, cfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)
}

func (s *baseIntegrationTestSuite) createAppRunnerConfig(appID string) {
	// create app runner config
	runnerCfgReq := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
	runnerCfgReq.Type = models.AppAppRunnerTypeAwsDashEcs

	runnerCfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, appID, runnerCfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), runnerCfg)
}

func (s *baseIntegrationTestSuite) createApp() *models.AppApp {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)

	s.createAppSandboxConfig(app.ID)
	s.createAppRunnerConfig(app.ID)
	return app
}

func (s *baseIntegrationTestSuite) formatInterpolatedString(val string) string {
	val = strings.ReplaceAll(val, " ", "_")
	return strings.ToLower(val)
}

func (s *baseIntegrationTestSuite) formatInputs(inputs map[string]models.ServiceAppInputRequest) map[string]models.ServiceAppInputRequest {
	formattedInputs := make(map[string]models.ServiceAppInputRequest, len(inputs))
	for k, input := range inputs {
		formattedK := s.formatInterpolatedString(k)
		formattedInputs[formattedK] = input
	}
	return formattedInputs
}

func (s *baseIntegrationTestSuite) createComponent(appID string) *models.AppComponent {
	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	compReq.Name = generics.ToPtr(s.formatInterpolatedString(*compReq.Name))
	compReq.Dependencies = []string{}

	comp, err := s.apiClient.CreateComponent(s.ctx, appID, compReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), comp)
	return comp
}

func (s *baseIntegrationTestSuite) createInstall(appID string) *models.AppInstall {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	fakeReq.Inputs = nil

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
