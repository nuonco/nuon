package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v50/github"
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

func (s *VCSServiceTestSuite) TestBuildStatusResponseWrapsSuspendedUser() {
	now := time.Now().UTC()
	login := "octocat"
	id := int64(123)

	resp := buildStatusResponse(&github.Installation{
		SuspendedAt: &github.Timestamp{Time: now},
		SuspendedBy: &github.User{Login: &login, ID: &id},
	}, "12345", now)

	require.NotNil(s.T(), resp.SuspendedBy)
	assert.Equal(s.T(), login, resp.SuspendedBy.Login)
	assert.Equal(s.T(), id, resp.SuspendedBy.ID)
}
