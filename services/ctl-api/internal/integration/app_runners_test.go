package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type appRunnersSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppRunnersSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appRunnersSuite))
}

func (s *appRunnersSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appRunnersSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *appRunnersSuite) createApp() *models.AppApp {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)

	s.createAppSandboxConfig(app.ID)
	return app
}

func (s *appRunnersSuite) TestCreateAppRunnerConfig() {
	s.T().Run("successfully created", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		cfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppRunnerLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)
		require.Equal(t, latestCfg.EnvVars, cfg.EnvVars)
		require.Equal(t, latestCfg.AppRunnerType, cfg.AppRunnerType)
	})

	s.T().Run("updates installs to reference new runner", func(t *testing.T) {
		install := s.createInstall(s.appID)

		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		appRunnerCfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, appRunnerCfg)

		updatedInstall, err := s.apiClient.GetInstall(s.ctx, install.ID)
		require.NoError(t, err)
		require.NotEmpty(t, updatedInstall)
		require.Equal(t, updatedInstall.AppRunnerConfig.ID, appRunnerCfg.ID)
	})

	s.T().Run("errors on invalid app id", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		appRunnerCfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, generics.GetFakeObj[string](), req)
		require.Error(t, err)
		require.Nil(t, appRunnerCfg)
		require.True(t, nuon.IsBadRequest(err))
	})
}

func (s *appRunnersSuite) TestGetAppRunnerLatestConfig() {
	s.T().Run("error when no configs found", func(t *testing.T) {
		_, err := s.apiClient.GetAppRunnerLatestConfig(s.ctx, s.appID)
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("success", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		cfg, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppRunnerLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)
		require.Equal(t, latestCfg.EnvVars, cfg.EnvVars)
		require.Equal(t, latestCfg.AppRunnerType, cfg.AppRunnerType)
	})

	s.T().Run("returns correct one when multiple", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		cfg1, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg1)

		cfg2, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg2)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppRunnerLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)
		require.Equal(t, latestCfg.ID, cfg2.ID)
	})

	s.T().Run("errors when invalid app", func(t *testing.T) {
		latestCfg, err := s.apiClient.GetAppRunnerLatestConfig(s.ctx, generics.GetFakeObj[string]())
		require.Nil(t, latestCfg)
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})

}

func (s *appRunnersSuite) TestGetAppRunnerConfigs() {
	s.T().Run("success", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		cfg1, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg1)

		cfgs, err := s.apiClient.GetAppRunnerConfigs(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, cfgs)
	})

	s.T().Run("returns in correct order", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppRunnerConfigRequest]()
		req.Type = models.AppAppRunnerTypeAwsDashEcs

		cfg1, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg1)

		cfg2, err := s.apiClient.CreateAppRunnerConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg2)

		cfgs, err := s.apiClient.GetAppRunnerConfigs(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, cfgs)
		require.Equal(t, cfgs[0].ID, cfg2.ID)
	})
}
