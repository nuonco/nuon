package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type installComponentsTestSuite struct {
	baseIntegrationTestSuite

	orgID     string
	appID     string
	compID    string
	buildID   string
	installID string
}

func TestInstallComponentsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installComponentsTestSuite))
}

func (s *installComponentsTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *installComponentsTestSuite) SetupTest() {
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID

	// create a component
	comp := s.createComponent(s.appID)
	s.compID = comp.ID

	// create a component config
	req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	// create a build of this component
	buildReq := &models.ServiceCreateComponentBuildRequest{
		GitRef: "HEAD",
	}
	build, err := s.apiClient.CreateComponentBuild(s.ctx, comp.ID, buildReq)
	require.NoError(s.T(), err)
	s.buildID = build.ID

	// create install
	install := s.createInstall(s.appID)
	s.installID = install.ID
}

func (s *installComponentsTestSuite) TestGetInstallComponents() {
	s.T().Run("get install components", func(t *testing.T) {
		installComponents, _, err := s.apiClient.GetInstallComponents(s.ctx, s.installID, nil)
		require.NoError(t, err)
		require.Len(t, installComponents, 1)
		require.Equal(t, s.compID, installComponents[0].ComponentID)
	})

	s.T().Run("returns components based on created order desc", func(t *testing.T) {
		comp := s.createComponent(s.appID)

		installComponents, _, err := s.apiClient.GetInstallComponents(s.ctx, s.installID, nil)
		require.NoError(t, err)
		require.Len(t, installComponents, 2)
		require.Equal(t, comp.ID, installComponents[0].ComponentID)
	})

	s.T().Run("get install components invalid install", func(t *testing.T) {
		installComponents, _, err := s.apiClient.GetInstallComponents(s.ctx, generics.GetFakeObj[string](), nil)
		require.Error(t, err)
		require.Empty(t, installComponents)
	})
}

func (s *installComponentsTestSuite) TestGetInstallComponentDeploys() {
	s.T().Run("successfully returns when no deploys", func(t *testing.T) {
		installDeploys, _, err := s.apiClient.GetInstallComponentDeploys(s.ctx, s.installID, s.compID, nil)
		require.NoError(t, err)
		require.Empty(t, installDeploys)
	})

	s.T().Run("success", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: s.buildID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), deploy)

		installDeploys, _, err := s.apiClient.GetInstallComponentDeploys(s.ctx, s.installID, s.compID, nil)
		require.NoError(t, err)
		require.Len(t, installDeploys, 1)
		require.Equal(t, installDeploys[0].ID, deploy.ID)
	})

	s.T().Run("sorts deploys based on created at DESC", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: s.buildID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), deploy)

		installDeploys, _, err := s.apiClient.GetInstallComponentDeploys(s.ctx, s.installID, s.compID, nil)
		require.NoError(t, err, "HELLO WORLD")
		require.Len(t, installDeploys, 2)
		require.Equal(t, installDeploys[0].ID, deploy.ID)
	})

	s.T().Run("errors on invalid install", func(t *testing.T) {
		installComponents, _, err := s.apiClient.GetInstallComponentDeploys(s.ctx, generics.GetFakeObj[string](), s.compID, nil)
		require.Error(t, err)
		require.Empty(t, installComponents)
	})

	s.T().Run("errors on invalid component", func(t *testing.T) {
		installComponents, _, err := s.apiClient.GetInstallComponentDeploys(s.ctx, s.installID, generics.GetFakeObj[string](), nil)
		require.Error(t, err)
		require.Empty(t, installComponents)
	})
}

func (s *installComponentsTestSuite) TestGetInstallComponentLatestDeploy() {
	s.T().Run("errors when no deploys", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallComponentLatestDeploy(s.ctx, s.installID, s.compID)
		require.Error(t, err)
		require.Nil(t, deploy)
	})

	s.T().Run("success", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: s.buildID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), deploy)

		installDeploy, err := s.apiClient.GetInstallComponentLatestDeploy(s.ctx, s.installID, s.compID)
		require.NoError(t, err)
		require.NotNil(t, installDeploy)
		require.Equal(t, installDeploy.ID, deploy.ID)
	})
}

func (s *installComponentsTestSuite) TestTeardownComponents() {
	s.T().Run("success", func(t *testing.T) {
		err := s.apiClient.TeardownInstallComponents(s.ctx, s.installID)
		require.NoError(t, err)
	})

	s.T().Run("install not found", func(t *testing.T) {
		err := s.apiClient.TeardownInstallComponents(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *installComponentsTestSuite) TestDeployComponents() {
	s.T().Run("success", func(t *testing.T) {
		err := s.apiClient.DeployInstallComponents(s.ctx, s.installID, false)
		require.NoError(t, err)
	})

	s.T().Run("install not found", func(t *testing.T) {
		err := s.apiClient.DeployInstallComponents(s.ctx, generics.GetFakeObj[string](), false)
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})
}
