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

type vcsIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID string
}

func TestVCSSuite(t *testing.T) {
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
	org := s.createOrg()
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
		require.Error(t, err)
		require.Nil(t, org)
	})
}

func (s *vcsIntegrationTestSuite) TestCreateConnectionCallback() {
	s.apiClient.SetOrgID("")

	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionCallbackRequest]()
		vcsReq.OrgID = generics.ToPtr(s.orgID)

		vcs, err := s.apiClient.CreateVCSConnectionCallback(s.ctx, vcsReq)
		require.Nil(t, err)
		require.NotNil(t, vcs)
		require.Equal(t, vcs.GithubInstallID, *(vcsReq.GithubInstallID))
	})

	s.T().Run("bad request", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateConnectionCallbackRequest]()
		req.OrgID = nil

		conn, err := s.apiClient.CreateVCSConnectionCallback(s.ctx, req)
		require.Error(t, err)
		require.Nil(t, conn)
		require.True(t, nuon.IsBadRequest(err))
	})

	s.T().Run("org not found", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateConnectionCallbackRequest]()
		req.OrgID = generics.GetFakeObj[*string]()

		conn, err := s.apiClient.CreateVCSConnectionCallback(s.ctx, req)
		require.Error(t, err)
		require.Nil(t, conn)
		require.True(t, nuon.IsNotFound(err))
	})
}

func (s *vcsIntegrationTestSuite) TestGetConnections() {
	vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
	vcs, err := s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), vcs)

	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcs, _, err := s.apiClient.GetVCSConnections(s.ctx, nil)
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
		// add a currentVCS connection to the org
		currentVCS, err := s.apiClient.GetVCSConnection(s.ctx, vcs.ID)
		require.Nil(t, err)
		require.NotNil(t, currentVCS)

		require.Equal(t, currentVCS.GithubInstallID, *(vcsReq.GithubInstallID))
	})
}

// TODO: This test needs to be updated as the client has changed
// func (s *vcsIntegrationTestSuite) TestGetAllConnectedRepos() {
// 	s.T().Run("returns all connected repos", func(t *testing.T) {
// 		if s.githubInstallID == "" {
// 			t.Skip("skipping because INTEGRATION_GITHUB_INSTALL_ID is not set")
// 			return
// 		}

// 		repos, _, err := s.apiClient.GetAllVCSConnectedRepos(s.ctx, nil)  // func no longer exists
// 		require.NoError(t, err)
// 		require.NotEmpty(t, repos)

// 		found := false
// 		for _, repo := range repos {
// 			if *repo.Name == "mono" {
// 				found = true
// 				break
// 			}
// 		}
// 		require.True(t, found)
// 	})
// }
