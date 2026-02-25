package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepSuccess() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)),
		testseed.WithStepRetryable(true),
	)

	body := RetryWorkflowByIDRequest{
		StepID:    step.ID,
		Operation: RetryOperationRetryStep,
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/retry", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var result RetryWorkflowByIDResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Equal(s.T(), workflow.ID, result.WorkflowID)

	// Verify the step was marked as retried in the DB.
	var updatedStep app.WorkflowStep
	require.NoError(s.T(), s.deps.DB.Where("id = ?", step.ID).First(&updatedStep).Error)
	assert.True(s.T(), updatedStep.Retried, "step should be marked as retried")
}

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepSkipSuccess() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)),
		testseed.WithStepSkippable(true),
	)

	body := RetryWorkflowByIDRequest{
		StepID:    step.ID,
		Operation: RetryOperationSkipStep,
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/retry", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var result RetryWorkflowByIDResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Equal(s.T(), workflow.ID, result.WorkflowID)

	// Skip doesn't set Retried — only sends the rerun signal to the event loop.
	var updatedStep app.WorkflowStep
	require.NoError(s.T(), s.deps.DB.Where("id = ?", step.ID).First(&updatedStep).Error)
	assert.False(s.T(), updatedStep.Retried, "skip should not mark step as retried")
}

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepNotRetryable() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)),
		testseed.WithStepRetryable(false),
	)

	body := RetryWorkflowByIDRequest{
		StepID:    step.ID,
		Operation: RetryOperationRetryStep,
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/retry", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepNotInErrorState() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID,
		testseed.WithStepRetryable(true),
		// Status defaults to pending, not error.
	)

	body := RetryWorkflowByIDRequest{
		StepID:    step.ID,
		Operation: RetryOperationRetryStep,
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/retry", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	body := RetryWorkflowByIDRequest{
		StepID:    "nonexistent",
		Operation: RetryOperationRetryStep,
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/nonexistent/retry", workflow.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
