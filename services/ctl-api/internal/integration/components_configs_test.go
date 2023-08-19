package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
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
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentConfigsSuite))
}

func (s *componentConfigsSuite) SetupTest() {
	// create an org
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.apiClient.SetOrgID(org.ID)
	s.orgID = org.ID

	// add a vcs connection to the org
	vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
	_, err = s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
	require.Nil(s.T(), err)

	// create an app
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)
	s.appID = app.ID

	// create a component
	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), comp)
	s.compID = comp.ID
}

func (s *componentConfigsSuite) TestCreateDockerBuildComponentConfig() {
	s.T().Run("success with connected github config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
		req.ConnectedGithubVcsConfig = nil

		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("success with public config", func(t *testing.T) {
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
		req := generics.GetFakeObj[*models.ServiceCreateHelmComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateHelmComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		// assert the fields
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
	s.T().Run("success with connected github config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("success with public config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
		req.ConnectedGithubVcsConfig = nil

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

func (s *componentConfigsSuite) TestComponentConfigs() {
	s.T().Run("successfully returns one component config", func(t *testing.T) {
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
		req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotNil(t, latestCfg)

		require.Equal(t, cfg.ID, latestCfg.ExternalImage.ID)
	})

	s.T().Run("error on no configs", func(t *testing.T) {
		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), comp)

		latestCfg, err := s.apiClient.GetComponentLatestConfig(s.ctx, comp.ID)
		require.NotNil(t, err)
		require.Nil(t, latestCfg)
	})
}
