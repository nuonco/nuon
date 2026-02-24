package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *VCSServiceTestSuite) TestGetConnections_Success() {
	// Create a test connection
	s.createTestVCSConnection()

	rr := s.makeRequest(http.MethodGet, "/v1/vcs/connections", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var conns []*app.VCSConnection
	err := json.Unmarshal(rr.Body.Bytes(), &conns)
	require.NoError(s.T(), err)

	assert.GreaterOrEqual(s.T(), len(conns), 1, "Should return at least one connection")
	assert.Equal(s.T(), "12345", conns[0].GithubInstallID)
	assert.Equal(s.T(), "test-org", conns[0].GithubAccountName)
}
