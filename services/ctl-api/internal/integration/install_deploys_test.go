package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type installDeploysIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID     string
	appID     string
	compID    string
	buildID   string
	installID string
}

func TestInstallDeploysSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installDeploysIntegrationTestSuite))
}

func (s *installDeploysIntegrationTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *installDeploysIntegrationTestSuite) SetupTest() {
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

	// create a component config
	req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), cfg)

	// create a build of this component
	buildReq := &models.ServiceCreateComponentBuildRequest{
		GitRef: "HEAD",
	}
	build, err := s.apiClient.CreateComponentBuild(s.ctx, comp.ID, buildReq)
	require.NoError(s.T(), err)
	s.buildID = build.ID

	// create install
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), install)
	s.installID = install.ID
}

func (s *installDeploysIntegrationTestSuite) TestEnsureInstallComponent() {
	s.T().Run("should automatically have an install component to deploy too", func(t *testing.T) {
		installComps, err := s.apiClient.GetInstallComponents(s.ctx, s.installID)
		require.NoError(t, err)
		require.Len(t, installComps, 1)
		require.Equal(t, s.compID, installComps[0].Component.ID)
	})
}

func (s *installDeploysIntegrationTestSuite) TestCreateInstallDeploy() {
	s.T().Run("creates install deploy properly", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: s.buildID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.NoError(t, err)
		require.NotNil(t, deploy)
	})

	s.T().Run("errors when install is invalid", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: s.buildID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, "doesntexist", depReq)
		require.Error(t, err)
		require.Nil(t, deploy)
	})

	s.T().Run("errors when build is invalid", func(t *testing.T) {
		depReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: generics.GetFakeObj[string](),
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.Error(t, err)
		require.Nil(t, deploy)
	})
}

func (s *installDeploysIntegrationTestSuite) TestGetInstallDeploy() {
	depReq := &models.ServiceCreateInstallDeployRequest{
		BuildID: s.buildID,
	}
	seedDeploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedDeploy)

	s.T().Run("successfully fetches install deploy", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallDeploy(s.ctx, s.installID, seedDeploy.ID)
		require.NoError(t, err)
		require.Equal(t, deploy.ID, seedDeploy.ID)
	})

	s.T().Run("errors when install is invalid", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallDeploy(s.ctx, generics.GetFakeObj[string](), seedDeploy.ID)
		require.Error(t, err)
		require.Nil(t, deploy)
	})

	s.T().Run("errors when build is invalid", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallDeploy(s.ctx, s.installID, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, deploy)
	})
}

func (s *installDeploysIntegrationTestSuite) TestGetInstallDeploys() {
	depReq := &models.ServiceCreateInstallDeployRequest{
		BuildID: s.buildID,
	}
	seedDeploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedDeploy)

	s.T().Run("successfully fetches deploys", func(t *testing.T) {
		deploys, err := s.apiClient.GetInstallDeploys(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotEmpty(t, deploys)
		require.Equal(t, deploys[0].ID, seedDeploy.ID)
	})

	s.T().Run("successfully fetches with multiple components", func(t *testing.T) {
		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), comp)

		deploys, err := s.apiClient.GetInstallDeploys(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotEmpty(t, deploys)
		require.Equal(t, deploys[0].ID, seedDeploy.ID)
	})

	s.T().Run("successfully returns deploys in created_at desc order", func(t *testing.T) {
		secondDeploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), secondDeploy)

		deploys, err := s.apiClient.GetInstallDeploys(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotEmpty(t, deploys)
		require.Equal(t, deploys[0].ID, secondDeploy.ID)
	})
}

func (s *installDeploysIntegrationTestSuite) TestGetInstallLatestDeploy() {
	depReq := &models.ServiceCreateInstallDeployRequest{
		BuildID: s.buildID,
	}
	seedDeploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.installID, depReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedDeploy)

	s.T().Run("successfully fetches latest deploy", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallLatestDeploy(s.ctx, s.installID)
		require.NoError(t, err)
		require.Equal(t, deploy.ID, seedDeploy.ID)
	})

	s.T().Run("errors when install is invalid", func(t *testing.T) {
		deploy, err := s.apiClient.GetInstallLatestDeploy(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, deploy)
	})

	s.T().Run("errors when no deploy exists", func(t *testing.T) {
		// create install
		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), install)

		deploy, err := s.apiClient.GetInstallLatestDeploy(s.ctx, install.ID)
		require.Error(t, err)
		require.Nil(t, deploy)
	})
}

func (s *installDeploysIntegrationTestSuite) TestGetInstallDeployLogs() {
	s.T().Skip("deploy logs are not implemented yet")
}
