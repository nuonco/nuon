package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type appInputSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppInputSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appInputSuite))
}

func (s *appInputSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appInputSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *appInputSuite) TestCreateAppInputConfig() {
	s.T().Run("successfully creates app input config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
		req.Inputs = s.formatInputs(req.Inputs)
		resp, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	s.T().Run("errors on invalid app id", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
		resp, err := s.apiClient.CreateAppInputConfig(s.ctx, generics.GetFakeObj[string](), req)
		require.Error(t, err)
		require.Empty(t, resp)
	})
}

func (s *appInputSuite) TestGetAppLatestInputConfig() {
	s.T().Run("returns latest config", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
		req.Inputs = s.formatInputs(req.Inputs)
		_, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		resp, err := s.apiClient.GetAppInputLatestConfig(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	s.T().Run("errors on invalid app id", func(t *testing.T) {
		resp, err := s.apiClient.GetAppInputLatestConfig(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, resp)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *appInputSuite) TestGetAppInputConfigs() {
	s.T().Run("success when empty", func(t *testing.T) {
		cfgs, err := s.apiClient.GetAppInputConfigs(s.ctx, s.appID)
		require.NoError(t, err)
		require.Empty(t, cfgs)
	})

	s.T().Run("error on invalid app id", func(t *testing.T) {
		cfgs, err := s.apiClient.GetAppInputConfigs(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, cfgs)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("success with multiple configs", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
		req.Inputs = s.formatInputs(req.Inputs)
		cfg1, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		req = generics.GetFakeObj[*models.ServiceCreateAppInputConfigRequest]()
		req.Inputs = s.formatInputs(req.Inputs)
		cfg2, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		cfgs, err := s.apiClient.GetAppInputConfigs(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, cfgs)

		require.Len(t, cfgs, 2)
		require.Equal(t, cfgs[0].ID, cfg2.ID)
		require.Equal(t, cfgs[1].ID, cfg1.ID)
	})
}
