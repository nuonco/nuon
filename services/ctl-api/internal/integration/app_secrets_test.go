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

type appSecretSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppSecretsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appSecretSuite))
}

func (s *appSecretSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appSecretSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *appSecretSuite) TestCreateAppSecret() {
	s.T().Run("successfully creates an app secret", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		req.Name = generics.ToPtr(s.formatInterpolatedString(*req.Name))
		resp, err := s.apiClient.CreateAppSecret(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	s.T().Run("errors on invalid secret name", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		req.Name = generics.ToPtr(*req.Name + "-")

		resp, err := s.apiClient.CreateAppSecret(s.ctx, s.appID, req)
		require.Error(t, err)
		require.True(t, nuon.IsBadRequest(err))
		require.Empty(t, resp)
	})

	s.T().Run("errors on invalid app id", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		resp, err := s.apiClient.CreateAppSecret(s.ctx, generics.GetFakeObj[string](), req)
		require.Error(t, err)
		require.Empty(t, resp)
	})
}

func (s *appSecretSuite) TestDeleteAppSecret() {
	s.T().Run("successfully deletes an app secret", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		req.Name = generics.ToPtr(s.formatInterpolatedString(*req.Name))
		resp, err := s.apiClient.CreateAppSecret(s.ctx, s.appID, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		deleted, err := s.apiClient.DeleteAppSecret(s.ctx, s.appID, resp.ID)
		require.True(t, deleted)
		require.Empty(t, err)
	})
}

func (s *appSecretSuite) TestGetAppSecretConfigs() {
	s.T().Run("success when empty", func(t *testing.T) {
		cfgs, err := s.apiClient.GetAppSecrets(s.ctx, s.appID)
		require.NoError(t, err)
		require.Empty(t, cfgs)
	})

	s.T().Run("error on invalid app id", func(t *testing.T) {
		cfgs, err := s.apiClient.GetAppSecrets(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, cfgs)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("success with multiple secrets", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		req.Name = generics.ToPtr(s.formatInterpolatedString(*req.Name))

		secret, err := s.apiClient.CreateAppSecret(s.ctx, s.appID, req)
		require.NoError(t, err)

		req = generics.GetFakeObj[*models.ServiceCreateAppSecretRequest]()
		req.Name = generics.ToPtr(s.formatInterpolatedString(*req.Name))

		secret2, err := s.apiClient.CreateAppSecret(s.ctx, s.appID, req)
		require.NoError(t, err)

		secrets, err := s.apiClient.GetAppSecrets(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, secrets)

		require.Len(t, secrets, 2)
		require.Equal(t, secrets[0].ID, secret2.ID)
		require.Equal(t, secrets[1].ID, secret.ID)
	})
}
