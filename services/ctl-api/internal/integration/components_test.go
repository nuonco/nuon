package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type componentsSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestComponentsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentsSuite))
}

func (s *componentsSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *componentsSuite) SetupTest() {
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *componentsSuite) TestCreateComponent() {
	s.T().Run("success", func(t *testing.T) {
		createReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		createReq.Name = generics.ToPtr(s.formatInterpolatedString(*createReq.Name))
		createReq.VarName = s.formatInterpolatedString(createReq.VarName)

		createReq.Dependencies = []string{}
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, createReq)
		require.Nil(t, err)
		require.NotNil(t, comp)

		require.Equal(t, comp.Name, *(createReq.Name))
		require.Equal(t, comp.VarName, createReq.VarName)
	})

	s.T().Run("sets interpolation name as app name by default", func(t *testing.T) {
		createReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		createReq.Name = generics.ToPtr(s.formatInterpolatedString(*createReq.Name))
		createReq.VarName = ""

		createReq.Dependencies = []string{}
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, createReq)
		require.Nil(t, err)
		require.NotNil(t, comp)

		require.Equal(t, comp.Name, *(createReq.Name))
		require.Equal(t, comp.VarName, *(createReq.Name))
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, &models.ServiceCreateComponentRequest{})
		require.NotNil(t, err)
		require.Nil(t, comp)
	})

	s.T().Run("creates install components for prexisting installs", func(t *testing.T) {
		install := s.createInstall(s.appID)

		compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
		compReq.Name = generics.ToPtr(s.formatInterpolatedString(*compReq.Name))
		compReq.VarName = ""
		compReq.Dependencies = []string{}

		comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
		require.NoError(t, err)
		require.NotNil(t, comp)

		installComps, err := s.apiClient.GetInstallComponents(s.ctx, install.ID)
		require.Nil(t, err)
		require.NotEmpty(t, installComps)
	})
}

func (s *componentsSuite) TestUpdateComponent() {
	comp := s.createComponent(s.appID)

	s.T().Run("success", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateComponentRequest]()
		updateReq.Name = generics.ToPtr(s.formatInterpolatedString(*updateReq.Name))
		updateReq.VarName = s.formatInterpolatedString(updateReq.VarName)
		updateReq.Dependencies = []string{}
		updatedComp, err := s.apiClient.UpdateComponent(s.ctx, comp.ID, updateReq)

		require.Nil(t, err)
		require.Equal(t, updatedComp.Name, *(updateReq.Name))
		require.Equal(t, updatedComp.VarName, updateReq.VarName)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
		comp, err := s.apiClient.UpdateComponent(s.ctx, s.appID, &models.ServiceUpdateComponentRequest{})
		require.NotNil(t, err)
		require.Nil(t, comp)
	})
}

func (s *componentsSuite) TestDeleteComponent() {
	comp := s.createComponent(s.appID)

	s.T().Run("success", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteComponent(s.ctx, comp.ID)
		require.Nil(t, err)
		require.True(t, deleted)

		_, err = s.apiClient.GetComponent(s.ctx, comp.ID)
		require.NoError(t, err)
		// NOTE: the event loops delete the component, and change status. Do not test for that here, since it is
		// indeterministic.
	})

	s.T().Run("errors on not found", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteComponent(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.False(t, deleted)
	})
}

func (s *componentsSuite) TestGetAllComponents() {
	comp := s.createComponent(s.appID)

	s.T().Run("success with a single app", func(t *testing.T) {
		comps, err := s.apiClient.GetAllComponents(s.ctx)
		require.Nil(t, err)
		require.Len(t, comps, 1)
		require.Equal(t, comp.ID, comps[0].ID)
	})

	s.T().Run("success all apps ordered by component desc", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.NoError(s.T(), err)
		require.NotNil(s.T(), app)

		comp2 := s.createComponent(s.appID)

		comps, err := s.apiClient.GetAllComponents(s.ctx)
		require.Nil(t, err)
		require.Len(t, comps, 2)
		require.Equal(t, comp2.ID, comps[0].ID)
		require.Equal(t, comp.ID, comps[1].ID)
	})
}

func (s *componentsSuite) TestGetComponent() {
	comp := s.createComponent(s.appID)

	s.T().Run("success", func(t *testing.T) {
		fetched, err := s.apiClient.GetComponent(s.ctx, comp.ID)
		require.Nil(t, err)
		require.NotNil(t, fetched)
	})

	s.T().Run("success by name", func(t *testing.T) {
		fetched, err := s.apiClient.GetComponent(s.ctx, comp.Name)
		require.Nil(t, err)
		require.NotNil(t, fetched)
	})

	s.T().Run("error", func(t *testing.T) {
		fetched, err := s.apiClient.GetComponent(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, fetched)
	})
}

func (s *componentsSuite) TestGetAppComponents() {
	comp := s.createComponent(s.appID)

	s.T().Run("success", func(t *testing.T) {
		comps, err := s.apiClient.GetAppComponents(s.ctx, s.appID)
		require.Nil(t, err)
		require.Len(t, comps, 1)
		require.Equal(t, comp.ID, comps[0].ID)
	})
}
