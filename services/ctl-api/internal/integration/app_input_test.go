package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/powertoolsdev/mono/pkg/generics"
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
	s.T().Run("successfully creates app inputs and groups", func(t *testing.T) {
		req := s.fakeInputRequest()

		resp, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	s.T().Run("errors on missing group", func(t *testing.T) {
		req := s.fakeInputRequest()
		req.Groups = nil

		resp, err := s.apiClient.CreateAppInputConfig(s.ctx, generics.GetFakeObj[string](), req)

		require.Error(t, err)
		require.Empty(t, resp)
		require.True(t, nuon.IsBadRequest(err))
	})

	s.T().Run("errors on invalid app id", func(t *testing.T) {
		req := s.fakeInputRequest()
		resp, err := s.apiClient.CreateAppInputConfig(s.ctx, generics.GetFakeObj[string](), req)
		require.Error(t, err)
		require.Empty(t, resp)
	})
}

func (s *appInputSuite) TestGetAppLatestInputConfig() {
	s.T().Run("returns latest config", func(t *testing.T) {
		req := s.fakeInputRequest()
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
		req := s.fakeInputRequest()
		cfg1, err := s.apiClient.CreateAppInputConfig(s.ctx, s.appID, req)
		require.NoError(t, err)

		req = s.fakeInputRequest()
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
