package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nuonco/nuon/pkg/generics"
)

type appsTestSuite struct {
	baseIntegrationTestSuite

	orgID string
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
		require.Equal(t, app.Description, appReq.Description)
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

	s.T().Run("success by id", func(t *testing.T) {
		currentApp, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.NotNil(t, currentApp)
		require.Equal(t, app.ID, currentApp.ID)
		require.Equal(t, app.Name, currentApp.Name)
	})

	s.T().Run("success by name", func(t *testing.T) {
		currentApp, err := s.apiClient.GetApp(s.ctx, app.Name)
		require.Nil(t, err)
		require.NotNil(t, currentApp)
		require.Equal(t, app.ID, currentApp.ID)
		require.Equal(t, app.Name, currentApp.Name)
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

	s.T().Run("errors on non-existent app id", func(t *testing.T) {
		nonExistentID := "appnonexistent1234567890"
		app, err := s.apiClient.GetApp(s.ctx, nonExistentID)
		require.NotNil(t, err)
		require.Nil(t, app)
	})

	s.T().Run("errors on deleted app", func(t *testing.T) {
		// Create a new app
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		newApp, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, newApp)

		// Delete the app
		deleted, err := s.apiClient.DeleteApp(s.ctx, newApp.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		// Try to get the deleted app
		fetchedApp, err := s.apiClient.GetApp(s.ctx, newApp.ID)
		require.NotNil(t, err)
		require.Nil(t, fetchedApp)
	})

	s.T().Run("errors on app from different org", func(t *testing.T) {
		// Create a new org
		orgReq := s.fakeOrgRequest()
		org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		defer s.deleteOrg(org.ID)

		// Create app in new org
		s.apiClient.SetOrgID(org.ID)
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		otherApp, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, otherApp)

		// Switch back to original org
		s.apiClient.SetOrgID(s.orgID)

		// Try to get app from different org - should fail
		fetchedApp, err := s.apiClient.GetApp(s.ctx, otherApp.ID)
		require.NotNil(t, err)
		require.Nil(t, fetchedApp)
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
		require.Equal(t, updatedApp.Description, updateAppReq.Description)

		// fetch the app
		fetched, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, fetched.Name, updateAppReq.Name)
	})

	s.T().Run("success with partial update", func(t *testing.T) {
		originalApp, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.NotNil(t, originalApp)

		// Update only description
		updateAppReq := &models.ServiceUpdateAppRequest{
			Description: "Updated description only",
		}

		updatedApp, err := s.apiClient.UpdateApp(s.ctx, app.ID, updateAppReq)
		require.Nil(t, err)
		require.NotNil(t, updatedApp)
		require.Equal(t, updatedApp.Description, updateAppReq.Description)
		require.Equal(t, updatedApp.Name, originalApp.Name) // Name should remain unchanged
	})

	s.T().Run("errors on empty id", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()
		app, err := s.apiClient.UpdateApp(s.ctx, "", updateAppReq)
		require.Error(t, err)
		require.Nil(t, app)
	})

	s.T().Run("errors on invalid id", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()
		app, err := s.apiClient.UpdateApp(s.ctx, generics.GetFakeObj[string](), updateAppReq)
		require.Error(t, err)
		require.Nil(t, app)
	})

	s.T().Run("errors on non-existent app id", func(t *testing.T) {
		updateAppReq := generics.GetFakeObj[*models.ServiceUpdateAppRequest]()
		nonExistentID := "appnonexistent1234567890"
		app, err := s.apiClient.UpdateApp(s.ctx, nonExistentID, updateAppReq)
		require.Error(t, err)
		require.Nil(t, app)
	})
}

func (s *appsTestSuite) TestDeleteApp() {
	s.T().Run("success", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		deleted, err := s.apiClient.DeleteApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		// Verify the app was actually deleted
		fetched, err := s.apiClient.GetApp(s.ctx, app.ID)
		require.NotNil(t, err)
		require.Nil(t, fetched)
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

	s.T().Run("errors when deleting already deleted app", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		// First deletion should succeed
		deleted, err := s.apiClient.DeleteApp(s.ctx, app.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		// Second deletion should fail
		deleted, err = s.apiClient.DeleteApp(s.ctx, app.ID)
		require.NotNil(t, err)
		require.False(t, deleted)
	})

	s.T().Run("errors on non-existent app id", func(t *testing.T) {
		nonExistentID := "appnonexistent1234567890"
		deleted, err := s.apiClient.DeleteApp(s.ctx, nonExistentID)
		require.NotNil(t, err)
		require.False(t, deleted)
	})
}

func (s *appsTestSuite) TestGetApps() {
	s.T().Run("success with single app", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.Nil(t, err)
		require.NotNil(t, app)

		apps, _, err := s.apiClient.GetApps(s.ctx, nil)
		require.Nil(t, err)
		require.Len(t, apps, 1)
		require.Equal(t, app.ID, apps[0].ID)
	})

	s.T().Run("success with multiple apps", func(t *testing.T) {
		// Create first app
		appReq1 := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq1.Name = generics.ToPtr(s.formatInterpolatedString(*appReq1.Name))
		app1, err := s.apiClient.CreateApp(s.ctx, appReq1)
		require.Nil(t, err)
		require.NotNil(t, app1)

		// Create second app
		appReq2 := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq2.Name = generics.ToPtr(s.formatInterpolatedString(*appReq2.Name))
		app2, err := s.apiClient.CreateApp(s.ctx, appReq2)
		require.Nil(t, err)
		require.NotNil(t, app2)

		// Create third app
		appReq3 := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq3.Name = generics.ToPtr(s.formatInterpolatedString(*appReq3.Name))
		app3, err := s.apiClient.CreateApp(s.ctx, appReq3)
		require.Nil(t, err)
		require.NotNil(t, app3)

		// Fetch all apps
		apps, _, err := s.apiClient.GetApps(s.ctx, nil)
		require.Nil(t, err)
		require.Len(t, apps, 3)

		// Verify all app IDs are present
		appIDs := make(map[string]bool)
		for _, app := range apps {
			appIDs[app.ID] = true
		}
		require.True(t, appIDs[app1.ID])
		require.True(t, appIDs[app2.ID])
		require.True(t, appIDs[app3.ID])
	})

	s.T().Run("success with empty result set", func(t *testing.T) {
		// Create a new org with no apps
		orgReq := s.fakeOrgRequest()
		org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		defer s.deleteOrg(org.ID)

		// Set the new org context
		s.apiClient.SetOrgID(org.ID)
		defer s.apiClient.SetOrgID(s.orgID) // Reset to original org

		// Fetch apps for new org
		apps, _, err := s.apiClient.GetApps(s.ctx, nil)
		require.Nil(t, err)
		require.Len(t, apps, 0)
	})

	s.T().Run("success with pagination", func(t *testing.T) {
		// Create multiple apps
		for range 5 {
			appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
			appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
			_, err := s.apiClient.CreateApp(s.ctx, appReq)
			require.Nil(t, err)
		}

		// Fetch with pagination
		apps, hasMore, err := s.apiClient.GetApps(s.ctx, &models.GetPaginatedQuery{
			Limit: 2,
		})
		require.Nil(t, err)
		require.Len(t, apps, 2)
		require.True(t, hasMore)
	})

	s.T().Run("excludes deleted apps", func(t *testing.T) {
		// Create two apps
		appReq1 := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq1.Name = generics.ToPtr(s.formatInterpolatedString(*appReq1.Name))
		app1, err := s.apiClient.CreateApp(s.ctx, appReq1)
		require.Nil(t, err)
		require.NotNil(t, app1)

		appReq2 := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq2.Name = generics.ToPtr(s.formatInterpolatedString(*appReq2.Name))
		app2, err := s.apiClient.CreateApp(s.ctx, appReq2)
		require.Nil(t, err)
		require.NotNil(t, app2)

		// Delete first app
		deleted, err := s.apiClient.DeleteApp(s.ctx, app1.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		// Fetch all apps - should only return the non-deleted app
		apps, _, err := s.apiClient.GetApps(s.ctx, nil)
		require.Nil(t, err)
		require.Len(t, apps, 1)
		require.Equal(t, app2.ID, apps[0].ID)
	})
}
