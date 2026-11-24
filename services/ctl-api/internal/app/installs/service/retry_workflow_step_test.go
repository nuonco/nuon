package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	path := fmt.Sprintf("/v1/workflows/%s/steps/nonexistent/retry", workflow.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepRejectsAnotherWorkflow() {
	install := s.createTestInstall()
	requested := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	other := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), other.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)), testseed.WithStepRetryable(true))

	rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/workflows/%s/steps/%s/retry", requested.ID, step.ID), nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, "body: %s", rr.Body.String())
	require.Contains(s.T(), rr.Body.String(), "workflow step not found")
	_, _, err := s.installsService.mcpRetryStep(s.workflowMCPContext(), nil, mcpRetryStepInput{WorkflowID: requested.ID, StepID: step.ID})
	require.ErrorContains(s.T(), err, "workflow step not found")

	var unchanged app.WorkflowStep
	require.NoError(s.T(), s.deps.DB.First(&unchanged, "id = ?", step.ID).Error)
	require.Equal(s.T(), other.ID, unchanged.InstallWorkflowID)
	require.Equal(s.T(), app.StatusError, unchanged.Status.Status)
	require.False(s.T(), unchanged.Retried)
}
