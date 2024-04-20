package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type componentConfigsSuite struct {
	baseIntegrationTestSuite

	orgID  string
	appID  string
	compID string
}

func TestComponentConfigsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentConfigsSuite))
}

func (s *componentConfigsSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *componentConfigsSuite) SetupTest() {
	org := s.createOrg()
	s.orgID = org.ID
	app := s.createApp()
	s.appID = app.ID

	// create a component
	comp := s.createComponent(s.appID)
	s.compID = comp.ID
}

func (s *componentConfigsSuite) TestCreateDockerBuildComponentConfig() {
	s.T().Run("success with public git vcs config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("success with connected github vcs config", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, &models.ServiceCreateDockerBuildComponentConfigRequest{})
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})
}

func (s *componentConfigsSuite) TestCreateTerraformModuleComponentConfig() {
	s.T().Run("success with connected github config", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateTerraformModuleComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("success with public config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateTerraformModuleComponentConfigRequest]()
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(s.ctx, s.compID, &models.ServiceCreateTerraformModuleComponentConfigRequest{})
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})
}

func (s *componentConfigsSuite) TestCreateHelmComponentConfig() {
	s.T().Run("success with connected github config", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		require.Equal(t, cfg.Values, req.Values)
		require.Equal(t, cfg.ValuesFiles, req.ValuesFiles)
		require.Equal(t, cfg.ChartName, *req.ChartName)
	})

	s.T().Run("success with public config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, &models.ServiceCreateHelmComponentConfigRequest{})
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})
}

func (s *componentConfigsSuite) TestCreateExternalImageComponentConfig() {
	s.T().Run("success", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()

		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, &models.ServiceCreateExternalImageComponentConfigRequest{})
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})
}

func (s *componentConfigsSuite) TestCreateJobComponentConfig() {
	s.T().Run("success", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateJobComponentConfigRequest]()
		cfg, err := s.apiClient.CreateJobComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		cfg, err := s.apiClient.CreateJobComponentConfig(s.ctx, s.compID, &models.ServiceCreateJobComponentConfigRequest{})
		require.NotNil(t, err)
		require.Nil(t, cfg)
	})
}

func (s *componentConfigsSuite) TestComponentConfigs() {
	s.T().Run("successfully returns one component config", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")
		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		// assert that latest is this id
		cfgs, err := s.apiClient.GetComponentConfigs(s.ctx, s.compID)
		require.NoError(t, err)
		require.Len(t, cfgs, 1)
		require.Equal(t, cfgs[0].ID, cfg.ComponentConfigConnectionID)
	})

	s.T().Run("returns based on created at desc order", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")
		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		// assert that latest is this id
		cfgs, err := s.apiClient.GetComponentConfigs(s.ctx, s.compID)
		require.Nil(t, err)
		require.Len(t, cfgs, 2)
		require.Equal(t, cfgs[0].ID, cfg.ComponentConfigConnectionID)
	})
}

func (s *componentConfigsSuite) TestGetLatestComponentConfig() {
	s.T().Run("success with helm", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}

		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")
		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		// assert that latest is this id
		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.Helm.ID)
	})
	s.T().Run("success with terraform module", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}
		req := generics.GetFakeObj[*models.ServiceCreateTerraformModuleComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateTerraformModuleComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.TerraformModule.ID)
	})
	s.T().Run("success with docker build", func(t *testing.T) {
		if s.githubInstallID == "" {
			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
			return
		}
		req := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.DockerBuild.ID)
	})

	s.T().Run("success with external image", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()

		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.ExternalImage.ID)
	})

	s.T().Run("success with job", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateJobComponentConfigRequest]()

		cfg, err := s.apiClient.CreateJobComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.Job.ID)
	})

	s.T().Run("error on no configs", func(t *testing.T) {
		comp := s.createComponent(s.appID)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, comp.ID)
		require.NotNil(t, err)
		require.Nil(t, latestCfg)
	})
}
