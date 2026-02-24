package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestUpdateWorkflowSuccess() {
	result := s.createTestInstallViaAPI()
	require.NotEmpty(s.T(), result.WorkflowID)

	promptOpt := app.InstallApprovalOptionPrompt
	body := UpdateWorkflowRequest{
		ApprovalOption: &promptOpt,
	}

	path := fmt.Sprintf("/v1/workflows/%s", result.WorkflowID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *InstallsServiceTestSuite) TestUpdateWorkflowNotFound() {
	promptOpt := app.InstallApprovalOptionPrompt
	body := UpdateWorkflowRequest{
		ApprovalOption: &promptOpt,
	}

	rr := s.makeRequest(http.MethodPatch, "/v1/workflows/iwf_nonexistent_00000000", body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
