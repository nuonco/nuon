package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *VCSServiceTestSuite) TestCheckConnectionStatus_Success() {
	// Create a test connection
	conn := s.createTestVCSConnection()

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/vcs/connections/%s/check-status", conn.ID), nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var statusResp VCSConnectionStatusResponse
	err := json.Unmarshal(rr.Body.Bytes(), &statusResp)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "active", statusResp.Status)
	assert.Equal(s.T(), "12345", statusResp.GithubInstallID)
	assert.NotNil(s.T(), statusResp.Account)
	assert.Equal(s.T(), "test-org", statusResp.Account.Login)
	assert.NotEmpty(s.T(), statusResp.Permissions)
}
