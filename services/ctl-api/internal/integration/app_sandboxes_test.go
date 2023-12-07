package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type appSandboxesSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppSandboxesSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appSandboxesSuite))
}

func (s *appSandboxesSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appSandboxesSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp(s.orgID)
	s.appID = app.ID
}

func (s *appSandboxesSuite) TestCreateAppSandboxConfig() {
	s.T().Run("success with built in sandbox", func(t *testing.T) {
		sandbox, err := s.apiClient.GetSandbox(s.ctx, "aws-eks")
		require.NoError(t, err)
		require.NotEmpty(t, sandbox.Releases[0].ID)

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = sandbox.Releases[0].ID
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)
		require.Equal(t, latestCfg.SandboxReleaseID, sandbox.Releases[0].ID)
	})

	s.T().Run("updates installs to reference new sandbox", func(t *testing.T) {
		install := s.createInstall(s.appID)

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.ConnectedGithubVcsConfig = nil

		appSandboxCfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, appSandboxCfg)

		updatedInstall, err := s.apiClient.GetInstall(s.ctx, install.ID)
		require.NoError(t, err)
		require.NotEmpty(t, updatedInstall)
		require.Equal(t, updatedInstall.AppSandboxConfig.ID, appSandboxCfg.ID)
	})

	s.T().Run("errors on invalid built-in sandbox id", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = generics.GetFakeObj[string]()
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.Error(t, err)
		require.Empty(t, cfg)
	})

	s.T().Run("successfully stores public vcs config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)

		require.NotEmpty(t, latestCfg.PublicGitVcsConfig)
		require.Equal(t, latestCfg.PublicGitVcsConfig.Branch, *req.PublicGitVcsConfig.Branch)
		require.Equal(t, latestCfg.PublicGitVcsConfig.Directory, *req.PublicGitVcsConfig.Directory)
		require.Equal(t, latestCfg.PublicGitVcsConfig.Repo, *req.PublicGitVcsConfig.Repo)
	})

	s.T().Run("successfully stores connected github vcs config", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// grab latest and ensure it is correctly configured
		latestCfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, latestCfg)

		require.NotEmpty(t, latestCfg.ConnectedGithubVcsConfig)
		require.Equal(t, latestCfg.ConnectedGithubVcsConfig.Branch, req.ConnectedGithubVcsConfig.Branch)
		require.Equal(t, latestCfg.ConnectedGithubVcsConfig.Directory, *req.ConnectedGithubVcsConfig.Directory)
		require.Equal(t, latestCfg.ConnectedGithubVcsConfig.Repo, *req.ConnectedGithubVcsConfig.Repo)
	})

	s.T().Run("errors on invalid github repo format", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("mono")

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	s.T().Run("errors on forbidden github repo", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("some-other-user/mono")

		cfg, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}

func (s *appSandboxesSuite) TestGetAppSandboxLatestConfig() {
	s.T().Run("success with built in sandbox", func(t *testing.T) {
		sandbox, err := s.apiClient.GetSandbox(s.ctx, "aws-eks")
		require.NoError(t, err)
		require.NotEmpty(t, sandbox.Releases[0].ID)

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = sandbox.Releases[0].ID
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig = nil
		_, err = s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		cfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotEmpty(t, cfg.SandboxRelease)
		require.NotEmpty(t, cfg.SandboxRelease.Sandbox)
	})

	s.T().Run("success with connected github", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.PublicGitVcsConfig = nil
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")
		_, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		cfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotEmpty(t, cfg.ConnectedGithubVcsConfig)
	})

	s.T().Run("success with public vcs connection", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSandboxConfigRequest]()
		req.SandboxReleaseID = ""
		req.ConnectedGithubVcsConfig = nil
		_, err := s.apiClient.CreateAppSandboxConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		cfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.NotEmpty(t, cfg.PublicGitVcsConfig)
	})

	s.T().Run("no sandbox config found", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.NoError(t, err)
		require.NotNil(t, app)

		cfg, err := s.apiClient.GetAppSandboxLatestConfig(s.ctx, app.ID)
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}

func (s *appSandboxesSuite) TestGetAppSandboxConfigs() {
	s.T().Run("success", func(t *testing.T) {
		cfgs, err := s.apiClient.GetAppSandboxConfigs(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, cfgs)
	})
}
