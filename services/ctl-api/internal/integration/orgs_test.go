package integration

import (
	"os"
	"testing"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
		fakeReq.UseCustomCert = true

		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		require.NoError(t, err)
		require.NotNil(t, org)
		require.Equal(t, *fakeReq.Name, org.Name)
		require.True(t, org.SandboxMode)
		require.True(t, org.CustomCert)

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
		fetchedOrg, err := s.apiClient.GetOrg(s.ctx)
		require.NoError(t, err)
		require.Len(t, fetchedOrg.Users, 1)

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
		orgs, err := s.apiClient.GetOrgs(s.ctx)
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

func (s *orgsIntegrationTestSuite) TestCreateOrgUser() {
	fakeReq := s.fakeOrgRequest()

	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)
	defer s.deleteOrg(seedOrg.ID)

	email := generics.GetFakeObj[string]()
	user, err := s.intAPIClient.CreateAdminUser(s.ctx, email, time.Hour)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), user)

	s.T().Run("success", func(t *testing.T) {
		resp, err := s.apiClient.CreateOrgUser(s.ctx, &models.ServiceCreateOrgUserRequest{
			UserID: email,
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		fetchedOrg, err := s.apiClient.GetOrg(s.ctx)
		require.NoError(t, err)
		require.NotNil(t, fetchedOrg)
		require.Len(t, fetchedOrg.Users, 2)
	})
}

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
