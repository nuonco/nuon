package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type vcsIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID string
}

func TestVCSsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(vcsIntegrationTestSuite))
}

func (s *vcsIntegrationTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *vcsIntegrationTestSuite) SetupTest() {
	// create an org
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.apiClient.SetOrgID(org.ID)
	s.orgID = org.ID
}

func (s *vcsIntegrationTestSuite) TestCreateConnection() {
	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
		vcs, err := s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
		require.Nil(t, err)
		require.NotNil(t, vcs)
		require.Equal(t, vcs.GithubInstallID, *(vcsReq.GithubInstallID))
	})
	s.T().Run("invalid request", func(t *testing.T) {
		org, err := s.apiClient.CreateVCSConnection(s.ctx, &models.ServiceCreateConnectionRequest{})
		assert.Error(t, err)
		assert.Nil(t, org)
	})
}

func (s *vcsIntegrationTestSuite) TestGetConnections() {
	vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
	vcs, err := s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), vcs)

	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcs, err := s.apiClient.GetVCSConnections(s.ctx)
		require.Nil(t, err)
		require.NotNil(t, vcs)
	})
}

func (s *vcsIntegrationTestSuite) TestGetConnection() {
	vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
	vcs, err := s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), vcs)

	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcs, err := s.apiClient.GetVCSConnection(s.ctx, vcs.ID)
		require.Nil(t, err)
		require.NotNil(t, vcs)

		require.Equal(t, vcs.GithubInstallID, *(vcsReq.GithubInstallID))
	})
}

func (s *vcsIntegrationTestSuite) TestGetAllConnectedRepos() {
	s.T().Skip("skipping connected repos as it requires a live github install ID")
}
