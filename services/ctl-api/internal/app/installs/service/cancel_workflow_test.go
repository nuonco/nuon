package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestCancelWorkflowSuccess() {
	result := s.createTestInstallViaAPI()
	require.NotEmpty(s.T(), result.WorkflowID)

	path := fmt.Sprintf("/v1/workflows/%s/cancel", result.WorkflowID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusAccepted {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusAccepted, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCancelWorkflowNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/workflows/iwf_nonexistent_00000000/cancel", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
