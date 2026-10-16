package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

func (s *InstallsServiceTestSuite) TestCancelWorkflowWaitingForRunner() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeActionWorkflowRun)
	require.NoError(s.T(), s.deps.DB.Model(workflow).Update("status", app.NewCompositeStatus(s.ctx, app.StatusWaitingForRunner)).Error)

	rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/workflows/%s/cancel", workflow.ID), nil)
	require.Equal(s.T(), http.StatusAccepted, rr.Code, "body: %s", rr.Body.String())
	require.NoError(s.T(), s.deps.DB.Where(app.Workflow{ID: workflow.ID}).First(workflow).Error)
	require.Equal(s.T(), app.StatusCancelled, workflow.Status.Status)
}

func (s *InstallsServiceTestSuite) TestCancelWorkflowNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/workflows/iwf_nonexistent_00000000/cancel", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
