package service

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *VCSServiceTestSuite) TestCreateConnection_Success() {
	req := CreateConnectionRequest{
		GithubInstallID: "12345",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connections", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var conn app.VCSConnection
	err := s.service.DB.Where("org_id = ? AND github_install_id = ?", s.testOrg.ID, "12345").First(&conn).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testOrg.ID, conn.OrgID)
	assert.Equal(s.T(), "12345", conn.GithubInstallID)
	assert.Equal(s.T(), "12345", conn.GithubAccountID)
	assert.Equal(s.T(), "test-org", conn.GithubAccountName)
}

func (s *VCSServiceTestSuite) TestCreateConnection_InvalidRequest() {
	req := map[string]interface{}{} // Empty request missing required github_install_id

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connections", req)

	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
