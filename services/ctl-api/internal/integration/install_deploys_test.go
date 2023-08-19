package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type installDeploysIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestInstallDeploysSuite(t *testing.T) {
	// NOTE(jm): this suite isn't working until we flesh out the details of components
	return
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installsIntegrationTestSuite))
}

func (s *installDeploysIntegrationTestSuite) SetupTest() {
	// create an org
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.apiClient.SetOrgID(org.ID)
	s.orgID = org.ID

	// create an app
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.appID = app.ID
}

func (s *installDeploysIntegrationTestSuite) TestCreateInstallDeploy() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
	require.NoError(s.T(), err)

	compCfgReq := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
	cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, comp.ID, compCfgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cfg)

	buildReq := &models.ServiceCreateComponentBuildRequest{
		GitRef: "HEAD",
	}
	build, err := s.apiClient.CreateComponentBuild(s.ctx, comp.ID, buildReq)
	require.NoError(s.T(), err)

	s.T().Run("creates install deploy properly", func(t *testing.T) {
		compReq := &models.ServiceCreateInstallDeployRequest{
			BuildID: build.ID,
		}
		deploy, err := s.apiClient.CreateInstallDeploy(s.ctx, s.appID, compReq)
		require.NoError(t, err)
		require.NotNil(t, deploy)
	})
}
