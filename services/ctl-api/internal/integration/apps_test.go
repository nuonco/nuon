package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type appsTestSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appsTestSuite))
}

func (s *appsTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appsTestSuite) SetupTest() {
	// create an org
	orgReq := s.fakeOrgRequest()

	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.apiClient.SetOrgID(org.ID)
	s.orgID = org.ID
}

func (s *appsTestSuite) TestCreateApp() {
	s.T().Run("success", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		require.Equal(t, app.Name, *(appReq.Name))
		require.NotEmpty(t, app.ID)
	})

	s.T().Run("returns app sandbox", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		require.Equal(t, app.Name, *(appReq.Name))
		require.NotEmpty(t, app.ID)
	})

	s.T().Run("errors on duplicate name", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		require.Equal(t, app.Name, *(appReq.Name))
		require.NotEmpty(t, app.ID)

		dupeApp, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Error(t, err)
		require.Nil(t, dupeApp)
	})

	s.T().Run("allows creating with duplicate name after deleting", func(t *testing.T) {
		t.Skip("can not test for success after deleting duplicated name because objects are deleted by workers")
		return
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		deleted, err := s.apiClient.DeleteApp(s.ctx, app.ID)
		require.NoError(t, err)
		require.True(t, deleted)

		dupeApp, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.NoError(t, err)
		require.NotNil(t, dupeApp)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		app, err := s.apiClient.CreateApp(s.ctx, &models.ServiceCreateAppRequest{})
		require.NotNil(t, err)
		require.Nil(t, app)
	})
}

func (s *appsTestSuite) TestGetApp() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), app)

	s.T().Run("success", func(t *testing.T) {
		app, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.NotNil(t, app)
	})
	s.T().Run("success by name", func(t *testing.T) {
		app, err := s.apiClient.GetApp(s.ctx, app.Name)
		require.Nil(t, err)
		require.NotNil(t, app)
	})

	s.T().Run("errors on empty id", func(t *testing.T) {
		app, err := s.apiClient.GetApp(s.ctx, "")
		require.NotNil(t, err)
		require.Nil(t, app)
	})

	s.T().Run("errors on invalid id", func(t *testing.T) {
		app, err := s.apiClient.GetApp(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.Nil(t, app)
	})
}

func (s *appsTestSuite) TestUpdateApp() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), app)

	s.T().Run("success", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()

		updatedApp, err := s.apiClient.UpdateApp(s.ctx, app.ID, updateAppReq)
		require.Nil(t, err)
		require.NotNil(t, updatedApp)
		require.Equal(t, updatedApp.Name, updateAppReq.Name)

		// fetch the app
		fetched, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, fetched.Name, updateAppReq.Name)
	})

	s.T().Run("errors on empty id", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()
		app, err := s.apiClient.UpdateApp(s.ctx, "", updateAppReq)
		require.Error(t, err)
		require.Nil(t, app)
	})

	s.T().Run("errors on invalid id", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()
		app, err := s.apiClient.UpdateApp(s.ctx, "", updateAppReq)
		require.Error(t, err)
		require.Nil(t, app)
	})
}

func (s *appsTestSuite) TestDeleteApp() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), app)

	s.T().Run("success", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		// make sure the app was actually deleted
		//fetched, err := s.apiClient.GetApp(s.ctx, app.ID)
		//require.NotNil(t, err)
		//require.Nil(t, fetched)
	})

	s.T().Run("errors on empty id", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteApp(s.ctx, "")
		require.NotNil(t, err)
		require.False(t, deleted)
	})

	s.T().Run("errors on missing id", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteApp(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.False(t, deleted)
	})
}

func (s *appsTestSuite) TestGetApps() {
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), app)

	s.T().Run("success", func(t *testing.T) {
		apps, err := s.apiClient.GetApps(s.ctx)
		require.Nil(t, err)
		require.Len(t, apps, 1)
		require.Equal(t, app.ID, apps[0].ID)
	})
}
