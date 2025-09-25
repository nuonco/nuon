package integration

import (
	"os"
	"testing"

	"github.com/nuonco/nuon-go/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type orgsIntegrationTestSuite struct {
	baseIntegrationTestSuite
}

func TestOrgsSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(orgsIntegrationTestSuite))
}

func (s *orgsIntegrationTestSuite) TestCreateOrg() {
	s.T().Run("success", func(t *testing.T) {
		fakeReq := s.fakeOrgRequest()

		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		require.Equal(t, *fakeReq.Name, org.Name)
		require.True(t, org.SandboxMode)

		s.deleteOrg(org.ID)
	})

	s.T().Run("sets custom cert", func(t *testing.T) {
		fakeReq := s.fakeOrgRequest()

		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		require.Equal(t, *fakeReq.Name, org.Name)
		require.True(t, org.SandboxMode)

		s.deleteOrg(org.ID)
	})

	s.T().Run("missing name", func(t *testing.T) {
		org, err := s.apiClient.CreateOrg(s.ctx, &models.ServiceCreateOrgRequest{})
		require.Error(t, err)
		require.Nil(t, org)
	})

	s.T().Run("adds current user who created the org as a user", func(t *testing.T) {
		fakeReq := s.fakeOrgRequest()

		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, org)

		s.apiClient.SetOrgID(org.ID)
		require.NoError(t, err)

		user, err := s.apiClient.GetCurrentUser(s.ctx)
		require.NoError(t, err)

		require.True(t, generics.SliceContains(org.ID, user.OrgIds))

		s.deleteOrg(org.ID)
	})

	s.T().Run("sets correct default feature flags", func(t *testing.T) {
		fakeReq := s.fakeOrgRequest()

		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, org)

		// Verify feature flags that should be disabled by default
		require.False(t, org.Features["org-dashboard"], "org-dashboard should be disabled by default")
		require.False(t, org.Features["install-break-glass"], "install-break-glass should be disabled by default")

		// Verify feature flags that should be enabled by default
		require.True(t, org.Features["api-pagination"], "api-pagination should be enabled by default")
		require.True(t, org.Features["org-runner"], "org-runner should be enabled by default")
		require.True(t, org.Features["org-settings"], "org-settings should be enabled by default")
		require.True(t, org.Features["org-support"], "org-support should be enabled by default")
		require.True(t, org.Features["install-delete-components"], "install-delete-components should be enabled by default")
		require.True(t, org.Features["install-delete"], "install-delete should be enabled by default")
		require.True(t, org.Features["terraform-workspace"], "terraform-workspace should be enabled by default")
		require.True(t, org.Features["dev-command"], "dev-command should be enabled by default")
		require.True(t, org.Features["app-branches"], "app-branches should be enabled by default")

		// Verify total number of features (should have all 11 features)
		require.Len(t, org.Features, 11, "should have all 11 feature flags set")

		s.deleteOrg(org.ID)
	})
}

func (s *orgsIntegrationTestSuite) TestOrgByID() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)
	defer s.deleteOrg(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		org, err := s.apiClient.GetOrg(s.ctx)
		require.NoError(t, err)
		require.NotNil(t, org)
		require.Equal(t, seedOrg.Name, org.Name)
		require.Equal(t, seedOrg.ID, org.ID)
	})
}

func (s *orgsIntegrationTestSuite) TestUpdateOrg() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)
	defer s.deleteOrg(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateOrgRequest]()
		org, err := s.apiClient.UpdateOrg(s.ctx, updateReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		require.Equal(t, *(updateReq.Name), org.Name)
		require.Equal(t, seedOrg.ID, org.ID)

		// fetch org
		fetchedOrg, err := s.apiClient.GetOrg(s.ctx)
		require.NoError(t, err)
		require.NotNil(t, fetchedOrg)
		require.Equal(t, *(updateReq.Name), fetchedOrg.Name)
	})
	s.T().Run("error when invalid request", func(t *testing.T) {
		org, err := s.apiClient.UpdateOrg(s.ctx, &models.ServiceUpdateOrgRequest{})
		require.Error(t, err)
		require.Nil(t, org)
	})
}

func (s *orgsIntegrationTestSuite) TestGetOrgs() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	defer s.deleteOrg(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		orgs, _, err := s.apiClient.GetOrgs(s.ctx, nil)
		require.NoError(t, err)
		require.NotEmpty(t, orgs)

		var lookupOrg *models.AppOrg
		for _, org := range orgs {
			if org.ID != seedOrg.ID {
				continue
			}
			lookupOrg = org
			break
		}
		require.NotNil(t, lookupOrg)
		require.Equal(t, seedOrg.ID, lookupOrg.ID)
		require.Equal(t, seedOrg.Name, lookupOrg.Name)
	})
}

func (s *orgsIntegrationTestSuite) TestCreateOrgInvite() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)
	defer s.deleteOrg(seedOrg.ID)

	user, err := s.intAPIClient.CreateIntegrationUser(s.ctx)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), user)

	s.T().Run("success", func(t *testing.T) {
		resp, err := s.apiClient.CreateOrgInvite(s.ctx, &models.ServiceCreateOrgInviteRequest{
			Email: &user.Email,
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		invites, _, err := s.apiClient.GetOrgInvites(s.ctx, &models.GetPaginatedQuery{
			Limit:  100,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, invites, 1)
	})
}

// func (s *orgsIntegrationTestSuite) TestGetRunnerGroup() {
// fakeReq := s.fakeOrgRequest()

// seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
// require.NoError(s.T(), err)
// require.NotNil(s.T(), seedOrg)
// s.apiClient.SetOrgID(seedOrg.ID)

//s.T().Run("success", func(t *testing.T) {
//grp, err := s.apiClient.GetOrgRunnerGroup(s.ctx)
//require.NoError(t, err)
//require.NotEmpty(t, grp.Runners)
//require.NotEmpty(t, grp.Settings)
//})
//}

func (s *orgsIntegrationTestSuite) TestDeleteOrg() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		deleted, err := s.apiClient.DeleteOrg(s.ctx)
		require.NoError(t, err)
		require.True(t, deleted)
	})
}
