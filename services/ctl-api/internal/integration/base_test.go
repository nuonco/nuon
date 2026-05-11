package integration

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nuonco/nuon/pkg/api"
	"github.com/nuonco/nuon/pkg/generics"
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
		api.WithAdminEmail("integration@serviceaccount.nuon.co"),
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

	apiClient, err := nuon.New(
		nuon.WithValidator(s.v),
		nuon.WithAuthToken(intUser.APIToken),
		nuon.WithURL(apiURL),
	)
	require.NoError(s.T(), err)
	s.apiClient = apiClient

	s.githubInstallID = intUser.GithubInstallID
}

// uniqueName generates a unique name for test resources to enable parallel execution.
// Format: prefix-timestamp-random6chars (e.g., "test-org-1709123456-a3f5k2")
func (s *baseIntegrationTestSuite) uniqueName(prefix string) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d-%s", prefix, timestamp, string(b))
}

func (s *baseIntegrationTestSuite) fakeOrgRequest() *models.ServiceCreateOrgRequest {
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	orgReq.UseSandboxMode = true
	return orgReq
}

func (s *baseIntegrationTestSuite) createOrg() *models.AppOrg {
	orgReq := s.fakeOrgRequest()
	name := s.uniqueName("test-org")
	orgReq.Name = &name

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

func (s *baseIntegrationTestSuite) fakeInstallInputsForApp(appID string) map[string]string {
	inputCfg, err := s.apiClient.GetAppInputLatestConfig(s.ctx, appID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), inputCfg)

	vals := make(map[string]string, 0)
	for _, input := range inputCfg.Inputs {
		vals[input.Name] = generics.GetFakeObj[string]()
	}

	return vals
}

func (s *baseIntegrationTestSuite) createAppWithInputs() *models.AppApp {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	name := s.formatInterpolatedString(*appReq.Name)
	appReq.Name = generics.ToPtr(s.uniqueName(name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)

	s.createAppSandboxConfig(app.ID)
	s.createAppRunnerConfig(app.ID)

	inputReq := s.fakeInputRequest()
	cfg, err := s.apiClient.CreateAppInputConfig(s.ctx, app.ID, inputReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	return app
}

func (s *baseIntegrationTestSuite) createAppSandboxConfig(appID string) {
	appConfigID := s.ensureAppConfig(appID)

	cfgReq := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
	cfgReq.ConnectedGithubVcsConfig = nil
	cfgReq.OperationRoles = nil
	cfgReq.AppConfigID = appConfigID

	cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, appID, cfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)
}

func (s *baseIntegrationTestSuite) ensureAppConfig(appID string) string {
	parentApp, err := s.apiClient.GetApp(s.ctx, appID)
	require.NoError(s.T(), err)
	if len(parentApp.AppConfigs) > 0 {
		return parentApp.AppConfigs[0].ID
	}

	cfgReq := generics.GetFakeObj[*models.ServiceCreateAppConfigRequest]()
	cfgReq.CliVersion = "0.19.922"
	cfg, err := s.apiClient.CreateAppConfig(s.ctx, appID, cfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	_, err = s.apiClient.UpdateAppConfig(s.ctx, appID, cfg.ID, &models.ServiceUpdateAppConfigRequest{
		Status: models.AppAppConfigStatusActive,
	})
	require.NoError(s.T(), err)
	return cfg.ID
}

func (s *baseIntegrationTestSuite) fakeRunnerConfigReq(appID string) *models.ServiceCreateAppRunnerConfigRequest {
	req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
	req.Type = models.NewAppAppRunnerType(models.AppAppRunnerTypeAwsDashEcs)
	req.AppConfigID = s.ensureAppConfig(appID)
	return req
}

func (s *baseIntegrationTestSuite) fakeSandboxConfigReq(appID string) *models.ServiceCreateAppSandboxConfigRequest {
	req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
	req.ConnectedGithubVcsConfig = nil
	req.OperationRoles = nil
	req.AppConfigID = s.ensureAppConfig(appID)
	return req
}

func (s *baseIntegrationTestSuite) createAppRunnerConfig(appID string) {
	appConfigID := s.ensureAppConfig(appID)

	runnerCfgReq := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
	runnerCfgReq.Type = models.NewAppAppRunnerType(models.AppAppRunnerTypeAwsDashEcs)
	runnerCfgReq.AppConfigID = appConfigID

	runnerCfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, appID, runnerCfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), runnerCfg)
}

func (s *baseIntegrationTestSuite) createApp() *models.AppApp {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	name := s.formatInterpolatedString(*appReq.Name)
	appReq.Name = generics.ToPtr(s.uniqueName(name))
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

func (s *baseIntegrationTestSuite) fakeInputRequest() *models.ServiceCreateAppInputConfigRequest {
	req := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
	req.Inputs = s.formatInputs(req.Inputs)

	for k, input := range req.Inputs {
		input.Type = "string"
		req.Inputs[k] = input
		req.Groups[*input.Group] = models.ServiceAppGroupRequest{
			Description: generics.GetFakeObj[*string](),
			DisplayName: generics.GetFakeObj[*string](),
		}
	}

	return req
}

func (s *baseIntegrationTestSuite) createComponent(appID string) *models.AppComponent {
	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	name := strings.ReplaceAll(s.uniqueName(s.formatInterpolatedString(*compReq.Name)), "-", "_")
	compReq.Name = generics.ToPtr(name)
	compReq.VarName = s.formatInterpolatedString(compReq.VarName)
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
	fakeReq.InstallConfig = nil

	install, err := s.apiClient.CreateInstall(s.ctx, appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), install)

	return install
}

func (s *baseIntegrationTestSuite) deleteOrg(orgID string) {
	// Skip cleanup in CI or when explicitly disabled
	// CI databases are ephemeral and unique names prevent collisions
	if os.Getenv("CI") != "" || os.Getenv("INTEGRATION_NO_CLEANUP") != "" {
		return
	}

	err := s.intAPIClient.DeleteOrg(s.ctx, orgID)
	require.NoError(s.T(), err)
}
