package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type componentBuildsSuite struct {
	baseIntegrationTestSuite

	orgID           string
	appID           string
	compID          string
	cfgID           string
	cfgConnectionID string
}

func TestComponentBuildsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentBuildsSuite))
}

func (s *componentBuildsSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *componentBuildsSuite) SetupTest() {
	// create an org
	orgReq := s.fakeOrgRequest()
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
	comp := s.createComponent(s.appID)
	s.compID = comp.ID

	// create a component config
	req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), cfg)
	s.cfgID = cfg.ID
	s.cfgConnectionID = cfg.ComponentConfigConnectionID
}

func (s *componentBuildsSuite) TestCreateComponentBuild() {
	s.T().Run("success with preset git ref", func(t *testing.T) {
		bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
		bldReq.GitRef = "head"
		bldReq.UseLatest = false

		bld, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
		require.Nil(t, err)
		require.NotNil(t, bld)
		require.NotEmpty(t, bld.ID)

		// make sure it creates the build for the correct component config
		require.Equal(t, bld.ComponentConfigConnectionID, s.cfgConnectionID)
	})

	s.T().Run("errors on invalid component id", func(t *testing.T) {
		bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
		bldReq.GitRef = "head"
		bldReq.UseLatest = false

		bld, err := s.apiClient.CreateComponentBuild(s.ctx, generics.GetFakeObj[string](), bldReq)
		require.NotNil(t, err)
		require.Nil(t, bld)
	})

	s.T().Run("errors when no component config is set", func(t *testing.T) {
		comp := s.createComponent(s.appID)

		bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
		bldReq.GitRef = "head"
		bldReq.UseLatest = false

		bld, err := s.apiClient.CreateComponentBuild(s.ctx, comp.ID, bldReq)
		require.NotNil(t, err)
		require.Nil(t, bld)
	})

	s.T().Run("successfully creates git commit", func(t *testing.T) {
		t.Skip("skipping vcs git configs to prevent coupling to our install id")
	})
}

func (s *componentBuildsSuite) TestGetComponentBuilds() {
	bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
	bldReq.GitRef = "head"
	bldReq.UseLatest = false
	bld, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), bld)

	s.T().Run("returns builds", func(t *testing.T) {
		blds, err := s.apiClient.GetComponentBuilds(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotEmpty(t, blds)
		require.Equal(t, blds[0].ID, bld.ID)
	})

	s.T().Run("errors on invalid component id", func(t *testing.T) {
		blds, err := s.apiClient.GetComponentBuilds(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.Empty(t, blds)
	})

	s.T().Run("returns in desc order by created at", func(t *testing.T) {
		secondBuild, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), secondBuild)

		blds, err := s.apiClient.GetComponentBuilds(s.ctx, s.compID)
		require.Nil(t, err)
		require.NotEmpty(t, blds)
		require.Len(t, blds, 2)

		// make sure it creates the build for the correct component config
		require.Equal(t, blds[0].ID, secondBuild.ID)
	})
}

func (s *componentBuildsSuite) TestGetComponentLatestBuild() {
	bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
	bldReq.GitRef = "head"
	bldReq.UseLatest = false
	bld, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), bld)

	s.T().Run("returns only build", func(t *testing.T) {
		latestBld, err := s.apiClient.GetComponentLatestBuild(s.ctx, s.compID)
		require.Nil(t, err)
		require.Equal(t, bld.ID, latestBld.ID)
	})

	s.T().Run("returns latest build", func(t *testing.T) {
		secondBuild, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), secondBuild)

		latestBld, err := s.apiClient.GetComponentLatestBuild(s.ctx, s.compID)
		require.Nil(t, err)
		require.Equal(t, secondBuild.ID, latestBld.ID)
	})

	s.T().Run("errors on invalid component id", func(t *testing.T) {
		bld, err := s.apiClient.GetComponentLatestBuild(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.Empty(t, bld)
	})

	s.T().Run("errors when no build exists", func(t *testing.T) {
		comp := s.createComponent(s.appID)

		// create a component config
		req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
		cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, comp.ID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)

		latestBld, err := s.apiClient.GetComponentLatestBuild(s.ctx, comp.ID)
		require.NotEmpty(t, err)
		require.Empty(t, latestBld)
	})
}

func (s *componentBuildsSuite) TestGetComponentBuild() {
	bldReq := generics.GetFakeObj[*models.ServiceCreateComponentBuildRequest]()
	bldReq.GitRef = "head"
	bldReq.UseLatest = false
	bld, err := s.apiClient.CreateComponentBuild(s.ctx, s.compID, bldReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), bld)

	s.T().Run("returns build", func(t *testing.T) {
		latestBld, err := s.apiClient.GetComponentBuild(s.ctx, s.compID, bld.ID)
		require.Nil(t, err)
		require.Equal(t, bld.ID, latestBld.ID)
	})

	s.T().Run("errors when build does not exist", func(t *testing.T) {
		returnedBld, err := s.apiClient.GetComponentBuild(s.ctx, s.compID, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, returnedBld)
	})
}

func (s *componentBuildsSuite) TestGetComponentBuildLogs() {
	s.T().Skip("not currently implemented")
}
