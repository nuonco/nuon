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

type installersSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestInstallersSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installersSuite))
}

func (s *installersSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *installersSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *installersSuite) TestCreateAppInstaller() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallerRequest]()
	fakeReq.AppIds = []string{s.appID}

	s.T().Run("success", func(t *testing.T) {
		installer, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, installer)

		require.Equal(t, installer.Type, models.AppInstallerTypeSelfHosted)
		require.Equal(t, installer.Apps[0].ID, s.appID)
		require.Equal(t, installer.Metadata.LogoURL, *fakeReq.Metadata.LogoURL)
		require.Equal(t, installer.Metadata.FaviconURL, fakeReq.Metadata.FaviconURL)
	})

	s.T().Run("failure with no app id", func(t *testing.T) {
		fakeReq.AppIds = nil
		installer, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)

		require.Error(t, err)
		require.True(t, nuon.IsBadRequest(err))
		require.Nil(t, installer)
	})

	s.T().Run("failure with invalid app id", func(t *testing.T) {
		fakeReq.AppIds = []string{generics.GetFakeObj[string]()}
		installer, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)

		require.Error(t, err)
		require.Nil(t, installer)
	})
}

func (s *installersSuite) TestGetAppInstaller() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallerRequest]()
	fakeReq.AppIds = []string{s.appID}
	inst, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)
	require.NoError(s.T(), err)

	s.T().Run("success", func(t *testing.T) {
		installer, err := s.apiClient.GetInstaller(s.ctx, inst.ID)
		require.NoError(t, err)
		require.NotEmpty(t, installer)
	})

	s.T().Run("error not found", func(t *testing.T) {
		installer, err := s.apiClient.GetInstaller(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, installer)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *installersSuite) TestUpdateAppInstaller() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallerRequest]()
	fakeReq.AppIds = []string{s.appID}
	inst, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)
	require.NoError(s.T(), err)

	s.T().Run("success", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateInstallerRequest]()
		updateReq.AppIds = fakeReq.AppIds

		installer, err := s.apiClient.UpdateInstaller(s.ctx, inst.ID, updateReq)
		require.NoError(t, err)
		require.NotEmpty(t, installer)
	})

	s.T().Run("error not found", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateInstallerRequest]()

		installer, err := s.apiClient.UpdateInstaller(s.ctx, generics.GetFakeObj[string](), updateReq)
		require.Error(t, err)
		require.Nil(t, installer)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("updates apps correctly", func(t *testing.T) {
		app := s.createApp()
		updateReq := generics.GetFakeObj[*models.ServiceUpdateInstallerRequest]()
		updateReq.AppIds = []string{app.ID}

		installer, err := s.apiClient.UpdateInstaller(s.ctx, inst.ID, updateReq)
		require.NoError(t, err)
		require.NotNil(t, installer)

		installer, err = s.apiClient.GetInstaller(s.ctx, inst.ID)
		require.NoError(t, err)
		require.Equal(t, 1, len(installer.Apps))
		require.Equal(t, installer.Apps[0].ID, app.ID)
	})
}

func (s *installersSuite) TestDeleteAppInstaller() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallerRequest]()
	fakeReq.AppIds = []string{s.appID}
	inst, err := s.apiClient.CreateInstaller(s.ctx, fakeReq)
	require.NoError(s.T(), err)

	s.T().Run("success", func(t *testing.T) {
		ok, err := s.apiClient.DeleteInstaller(s.ctx, inst.ID)
		require.NoError(t, err)
		require.True(t, ok)
	})

	s.T().Run("error not found", func(t *testing.T) {
		ok, err := s.apiClient.DeleteInstaller(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.False(t, ok)
		require.True(t, nuon.IsNotFound(err))
	})
}
