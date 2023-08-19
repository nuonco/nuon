package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type installComponentsTestSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestInstallComponentsSuite(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installComponentsTestSuite))
}

func (s *installComponentsTestSuite) SetupTest() {
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

func (s *installComponentsTestSuite) TestGetInstallComponents() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("creates install component when component exists first", func(t *testing.T) {
		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.NoError(t, err)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
		require.NoError(t, err)

		installComponents, err := s.apiClient.GetInstallComponents(s.ctx, install.ID)
		require.NoError(t, err)
		require.Len(t, installComponents, 1)
		require.Equal(t, installComponents[0].ComponentID, comp.ID)
	})

	s.T().Run("creates install component when component created after", func(t *testing.T) {
		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.NoError(t, err)

		installComponents, err := s.apiClient.GetInstallComponents(s.ctx, seedInstall.ID)
		require.NoError(t, err)
		require.Len(t, installComponents, 2)
		require.Equal(t, installComponents[1].ComponentID, comp.ID)
	})

	s.T().Run("get install components invalid install", func(t *testing.T) {
		installComponents, err := s.apiClient.GetInstallComponents(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, installComponents)
	})
}
