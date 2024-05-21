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
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *installsIntegrationTestSuite) TestCreateInstall() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	fakeReq.Inputs = nil

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
	s.T().Run("adding inputs", func(t *testing.T) {
		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()

		install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
		require.Error(t, err)
		require.True(t, nuon.IsBadRequest(err))
		require.Nil(t, install)
	})

	s.T().Run("adds existing components to install", func(t *testing.T) {
		comp := s.createComponent(s.appID)

		install := s.createInstall(s.appID)

		installComps, err := s.apiClient.GetInstallComponents(s.ctx, install.ID)
		require.NoError(t, err)
		require.Len(t, installComps, 1)
		require.Equal(t, installComps[0].ComponentID, comp.ID)
	})

	s.T().Run("errors when no app sandbox config exists", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.NoError(t, err)
		require.NotNil(t, app)

		s.createAppRunnerConfig(app.ID)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		install, err := s.apiClient.CreateInstall(s.ctx, app.ID, fakeReq)
		require.Error(t, err)
		require.Nil(t, install)
	})

	s.T().Run("errors when no app runner config exists", func(t *testing.T) {
		appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
		appReq.Name = generics.ToPtr(s.formatInterpolatedString(*appReq.Name))
		app, err := s.apiClient.CreateApp(s.ctx, appReq)
		require.NoError(t, err)
		require.NotNil(t, app)

		s.createAppSandboxConfig(app.ID)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"

		install, err := s.apiClient.CreateInstall(s.ctx, app.ID, fakeReq)
		require.Error(t, err)
		require.Nil(t, install)
	})

	s.T().Run("errors when app has inputs declared but are not provided", func(t *testing.T) {
		app := s.createAppWithInputs(s.orgID)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		fakeReq.Inputs = map[string]string{}

		install, err := s.apiClient.CreateInstall(s.ctx, app.ID, fakeReq)
		require.Error(t, err)
		require.Nil(t, install)
		require.True(t, nuon.IsBadRequest(err))
	})

	s.T().Run("errors install input is empty", func(t *testing.T) {
		app := s.createAppWithInputs(s.orgID)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		fakeReq.Inputs = s.fakeInstallInputsForApp(app.ID)
		for k := range fakeReq.Inputs {
			fakeReq.Inputs[k] = ""
		}

		install, err := s.apiClient.CreateInstall(s.ctx, app.ID, fakeReq)
		require.Error(t, err)
		require.Nil(t, install)
		require.True(t, nuon.IsBadRequest(err))
	})

	s.T().Run("successfully sets the inputs when valid", func(t *testing.T) {
		app := s.createAppWithInputs(s.orgID)

		fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
		fakeReq.AwsAccount.Region = "us-west-2"
		fakeReq.Inputs = s.fakeInstallInputsForApp(app.ID)

		install, err := s.apiClient.CreateInstall(s.ctx, app.ID, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, install)

		inputs, err := s.apiClient.GetInstallCurrentInputs(s.ctx, install.ID)
		require.NoError(t, err)
		require.NotNil(t, inputs)
	})
}

func (s *installsIntegrationTestSuite) TestGetInstall() {
	seedInstall := s.createInstall(s.appID)

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

func (s *installsIntegrationTestSuite) TestReprovisionInstall() {
	seedInstall := s.createInstall(s.appID)

	s.T().Run("success", func(t *testing.T) {
		err := s.apiClient.ReprovisionInstall(s.ctx, seedInstall.ID)
		require.NoError(t, err)
	})
	s.T().Run("invalid id", func(t *testing.T) {
		err := s.apiClient.ReprovisionInstall(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *installsIntegrationTestSuite) TestDeprovisionInstall() {
	seedInstall := s.createInstall(s.appID)

	s.T().Run("success", func(t *testing.T) {
		err := s.apiClient.DeprovisionInstall(s.ctx, seedInstall.ID)
		require.NoError(t, err)
	})
	s.T().Run("invalid id", func(t *testing.T) {
		err := s.apiClient.DeprovisionInstall(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *installsIntegrationTestSuite) TestDeleteInstall() {
	seedInstall := s.createInstall(s.appID)

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
	seedInstall := s.createInstall(s.appID)

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
	origInstall := s.createInstall(s.appID)

	s.T().Run("success", func(t *testing.T) {
		secondApp := s.createApp()
		s.createInstall(secondApp.ID)

		installs, err := s.apiClient.GetAppInstalls(s.ctx, s.appID)
		require.Nil(t, err)
		require.Len(t, installs, 1)
		require.Equal(t, installs[0].ID, origInstall.ID)
	})
	s.T().Run("errors when app not found", func(t *testing.T) {
		installs, err := s.apiClient.GetAppInstalls(s.ctx, generics.GetFakeObj[string]())
		require.NotNil(t, err)
		require.Empty(t, installs)
	})
}

func (s *installsIntegrationTestSuite) TestGetAllInstalls() {
	origInstall := s.createInstall(s.appID)

	secondApp := s.createApp()
	secondAppInstall := s.createInstall(secondApp.ID)

	s.T().Run("success", func(t *testing.T) {
		installs, err := s.apiClient.GetAllInstalls(s.ctx)
		require.Nil(t, err)
		require.Len(t, installs, 2)
		require.Equal(t, installs[0].ID, secondAppInstall.ID)
		require.Equal(t, installs[1].ID, origInstall.ID)
	})
}
