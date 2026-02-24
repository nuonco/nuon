package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *VCSServiceTestSuite) TestListConnectionRepos_Success() {
	// Create a test connection
	conn := s.createTestVCSConnection()

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/vcs/connections/%s/repos", conn.ID), nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var reposResp VCSConnectionReposResponse
	err := json.Unmarshal(rr.Body.Bytes(), &reposResp)
	require.NoError(s.T(), err)

	assert.Greater(s.T(), reposResp.TotalCount, 0, "Should return at least one repository")
	assert.NotEmpty(s.T(), reposResp.Repositories)
	assert.Equal(s.T(), "test-repo", reposResp.Repositories[0].Name)
	assert.Equal(s.T(), "test-org/test-repo", reposResp.Repositories[0].FullName)
}
