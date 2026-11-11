package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/generics"
)

type vcsIntegrationTestSuite struct {
	baseIntegrationTestSuite

	orgID string
}

func TestVCSSuite(t *testing.T) {
	t.Skip("TODO: faker-data validation issues; re-enable after test cleanup")

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
	s.createVCSConnection(s.orgID, *generics.GetFakeObj[*string]())

	s.T().Run("success", func(t *testing.T) {
		// add a vcs connection to the org
		vcs, _, err := s.apiClient.GetVCSConnections(s.ctx, nil)
		require.Nil(t, err)
		require.NotNil(t, vcs)
	})
}

func (s *vcsIntegrationTestSuite) TestGetConnection() {
	githubInstallID := *generics.GetFakeObj[*string]()
	vcs := s.createVCSConnection(s.orgID, githubInstallID)

	s.T().Run("success", func(t *testing.T) {
		currentVCS, err := s.apiClient.GetVCSConnection(s.ctx, vcs.ID)
		require.Nil(t, err)
		require.NotNil(t, currentVCS)

		require.Equal(t, githubInstallID, currentVCS.GithubInstallID)
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
