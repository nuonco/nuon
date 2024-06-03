package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type installInputsSuite struct {
	baseIntegrationTestSuite

	orgID     string
	appID     string
	installID string
}

func TestInstallInputsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(installInputsSuite))
}

func (s *installInputsSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *installInputsSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createAppWithInputs()
	s.appID = app.ID

	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	fakeReq.Inputs = s.fakeInstallInputsForApp(s.appID)

	install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), install)
	s.installID = install.ID
}

func (s *installInputsSuite) TestCreateInstallInputs() {
	s.T().Run("success with inputs", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateInstallInputsRequest]()
		req.Inputs = s.fakeInstallInputsForApp(s.appID)

		inputs, err := s.apiClient.CreateInstallInputs(s.ctx, s.installID, req)
		require.NoError(t, err)
		require.NotNil(t, inputs)
	})

	s.T().Run("errors on missing install inputs", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateInstallInputsRequest]()
		req.Inputs = map[string]string{}

		inputs, err := s.apiClient.CreateInstallInputs(s.ctx, s.installID, req)
		require.Error(t, err)
		require.Nil(t, inputs)
		require.True(t, nuon.IsBadRequest(err))
	})

	s.T().Run("errors on empty install inputs", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateInstallInputsRequest]()
		req.Inputs = s.fakeInstallInputsForApp(s.appID)
		for k := range req.Inputs {
			req.Inputs[k] = ""
		}

		inputs, err := s.apiClient.CreateInstallInputs(s.ctx, s.installID, req)
		require.Error(t, err)
		require.Nil(t, inputs)
		require.True(t, nuon.IsBadRequest(err))
	})
}

func (s *installInputsSuite) TestGetInstallInputs() {
	s.T().Run("success", func(t *testing.T) {
		inputs, err := s.apiClient.GetInstallInputs(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotEmpty(t, inputs)
		require.NotEmpty(t, inputs[0].Values)
	})

	s.T().Run("invalid install", func(t *testing.T) {
		inputs, err := s.apiClient.GetInstallInputs(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, inputs)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("ordering is correct", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateInstallInputsRequest]()
		req.Inputs = s.fakeInstallInputsForApp(s.appID)

		inputs, err := s.apiClient.CreateInstallInputs(s.ctx, s.installID, req)
		require.NoError(t, err)
		require.NotNil(t, inputs)

		allInputs, err := s.apiClient.GetInstallInputs(s.ctx, s.installID)
		require.NoError(t, err)
		require.Len(t, allInputs, 2)
		require.Equal(t, inputs.ID, allInputs[0].ID)
	})
}

func (s *installInputsSuite) TestGetInstallCurrentInputs() {
	s.T().Run("success", func(t *testing.T) {
		inputs, err := s.apiClient.GetInstallCurrentInputs(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotEmpty(t, inputs)
		require.NotEmpty(t, inputs.Values)
	})

	s.T().Run("invalid install", func(t *testing.T) {
		inputs, err := s.apiClient.GetInstallCurrentInputs(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, inputs)
		require.True(t, nuon.IsNotFound(err))
	})

	s.T().Run("ordering is correct", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateInstallInputsRequest]()
		req.Inputs = s.fakeInstallInputsForApp(s.appID)

		inputs, err := s.apiClient.CreateInstallInputs(s.ctx, s.installID, req)
		require.NoError(t, err)
		require.NotNil(t, inputs)

		currentInputs, err := s.apiClient.GetInstallCurrentInputs(s.ctx, s.installID)
		require.NoError(t, err)
		require.NotNil(t, currentInputs)
		require.Equal(t, inputs.ID, currentInputs.ID)
	})
}
