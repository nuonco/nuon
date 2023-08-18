package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type orgsIntegrationTestSuite struct {
	baseIntegrationTestSuite
}

func TestOrgsSuite(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(orgsIntegrationTestSuite))
}

func (s *orgsIntegrationTestSuite) TestCreateOrg() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()

	s.T().Run("success", func(t *testing.T) {
		org, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, *fakeReq.Name, org.Name)
	})
	s.T().Run("missing name", func(t *testing.T) {
		org, err := s.apiClient.CreateOrg(s.ctx, &models.ServiceCreateOrgRequest{})
		assert.Error(t, err)
		assert.Nil(t, org)
	})
}

func (s *orgsIntegrationTestSuite) TestOrgByID() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		org, err := s.apiClient.GetOrg(s.ctx, seedOrg.ID)
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, seedOrg.Name, org.Name)
		assert.Equal(t, seedOrg.ID, org.ID)
	})
	s.T().Run("error when org does not exist", func(t *testing.T) {
		org, err := s.apiClient.GetOrg(s.ctx, generics.GetFakeObj[string]())
		assert.Error(t, err)
		assert.Nil(t, org)
	})
	s.T().Run("errors with no org id", func(t *testing.T) {
		orgs, err := s.apiClient.GetOrg(s.ctx, seedOrg.ID)
		assert.Error(t, err)
		assert.Empty(t, orgs)
	})
}

func (s *orgsIntegrationTestSuite) TestUpdateOrg() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), seedOrg)
	s.apiClient.SetOrgID(seedOrg.ID)

	s.T().Run("success", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateOrgRequest]()
		org, err := s.apiClient.UpdateOrg(s.ctx, seedOrg.ID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, seedOrg.Name, org.Name)
		assert.Equal(t, seedOrg.ID, org.ID)

		// fetch org
		fetchedOrg, err := s.apiClient.GetOrg(s.ctx, seedOrg.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedOrg)
		assert.Equal(t, updateReq.Name, fetchedOrg.Name)
	})
	s.T().Run("error when org does not exist", func(t *testing.T) {
		updateReq := generics.GetFakeObj[*models.ServiceUpdateOrgRequest]()
		org, err := s.apiClient.UpdateOrg(s.ctx, generics.GetFakeObj[string](), updateReq)
		assert.Error(t, err)
		assert.Nil(t, org)
	})
	s.T().Run("error when invalid request", func(t *testing.T) {
		org, err := s.apiClient.UpdateOrg(s.ctx, seedOrg.ID, &models.ServiceUpdateOrgRequest{})
		assert.Error(t, err)
		assert.Nil(t, org)
	})
	s.T().Run("errors with no org id", func(t *testing.T) {
		s.apiClient.SetOrgID("")
		orgs, err := s.apiClient.UpdateOrg(s.ctx, seedOrg.ID, &models.ServiceUpdateOrgRequest{})
		assert.Error(t, err)
		assert.Empty(t, orgs)
	})
}

func (s *orgsIntegrationTestSuite) TestGetOrgs() {
	fakeReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	seedOrg, err := s.apiClient.CreateOrg(s.ctx, fakeReq)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), seedOrg)

	s.T().Run("success", func(t *testing.T) {
		orgs, err := s.apiClient.GetOrgs(s.ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, orgs)

		var lookupOrg *models.AppOrg
		for _, org := range orgs {
			if org.ID != seedOrg.ID {
				continue
			}
			lookupOrg = org
			break
		}
		assert.NotNil(t, lookupOrg)
		assert.Equal(t, seedOrg.ID, lookupOrg.ID)
		assert.Equal(t, seedOrg.Name, lookupOrg.Name)
	})
}
