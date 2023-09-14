package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type installsIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestInstallsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installsIntegrationTestSuite))
}

func (s *installsIntegrationTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *installsIntegrationTestSuite) SetupTest() {
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

func (s *installsIntegrationTestSuite) TestCreateInstall() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"

	s.T().Run("success", func(t *testing.T) {
		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, install)

		require.Equal(t, *fakeReq.Name, install.Name)
	})
	s.T().Run("missing name", func(t *testing.T) {
		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, &models.ServiceCreateInstallRequest{})
		require.Error(t, err)
		require.Nil(t, install)
	})

	s.T().Run("adds existing components to install", func(t *testing.T) {
		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.NoError(t, err)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, install)

		installComps, err := s.apiClient.GetInstallComponents(s.ctx, install.ID)
		require.NoError(t, err)
		require.Len(t, installComps, 1)
		require.Equal(t, installComps[0].ComponentID, comp.ID)
	})
}

func (s *installsIntegrationTestSuite) TestGetInstall() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("success", func(t *testing.T) {
		instl, err := s.apiClient.GetInstall(s.ctx, seedInstall.ID)
		require.Nil(t, err)
		require.NotNil(t, instl)
	})

	s.T().Run("success by name", func(t *testing.T) {
		instl, err := s.apiClient.GetInstall(s.ctx, seedInstall.Name)
		require.Nil(t, err)
		require.NotNil(t, instl)
	})

	s.T().Run("invalid id", func(t *testing.T) {
		install, err := s.apiClient.GetInstall(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, install)
	})
}

func (s *installsIntegrationTestSuite) TestDeleteInstall() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("success", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteInstall(s.ctx, seedInstall.ID)
		require.Nil(t, err)
		require.True(t, deleted)
	})
	s.T().Run("invalid id", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteInstall(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.False(t, deleted)
	})
}

func (s *installsIntegrationTestSuite) TestUpdateInstall() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("success", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateInstallRequest]()
		instl, err := s.apiClient.UpdateInstall(s.ctx, seedInstall.ID, updateReq)
		require.Nil(t, err)
		require.NotNil(t, instl)
		require.Equal(t, updateReq.Name, instl.Name)

		// fetch the install and verify it
		fetchedInstl, err := s.apiClient.GetInstall(s.ctx, seedInstall.ID)
		require.Nil(t, err)
		require.NotNil(t, fetchedInstl)
		require.Equal(t, updateReq.Name, fetchedInstl.Name)
	})
	s.T().Run("invalid id", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateInstallRequest]()
		install, err := s.apiClient.UpdateInstall(s.ctx, generics.GetFakeObj[string](), updateReq)
		require.Error(t, err)
		require.Nil(t, install)
	})
}

func (s *installsIntegrationTestSuite) TestGetAppInstalls() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)
	s.appID = app.ID

	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("success", func(t *testing.T) {
		installs, err := s.apiClient.GetAppInstalls(s.ctx, app.ID)
		require.Nil(t, err)
		require.Len(t, installs, 1)
		require.Equal(t, installs[0].ID, seedInstall.ID)
	})
	s.T().Run("errors when app not found", func(t *testing.T) {
		installs, err := s.apiClient.GetAppInstalls(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.Empty(t, installs)
	})
}

func (s *installsIntegrationTestSuite) TestGetAllInstalls() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)
	s.appID = app.ID

	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	seedInstall, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedInstall)

	s.T().Run("success", func(t *testing.T) {
		installs, err := s.apiClient.GetAllInstalls(s.ctx)
		require.Nil(t, err)
		require.Len(t, installs, 1)
		require.Equal(t, installs[0].ID, seedInstall.ID)
	})
}
