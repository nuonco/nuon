package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *VCSServiceTestSuite) TestGetConnection_Success() {
	// Create a test connection
	conn := s.createTestVCSConnection()

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/vcs/connections/%s", conn.ID), nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var respConn app.VCSConnection
	err := json.Unmarshal(rr.Body.Bytes(), &respConn)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), conn.ID, respConn.ID)
	assert.Equal(s.T(), "12345", respConn.GithubInstallID)
	assert.Equal(s.T(), "test-org", respConn.GithubAccountName)
}
