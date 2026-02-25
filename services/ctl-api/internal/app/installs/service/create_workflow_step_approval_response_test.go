package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestCreateApprovalResponseSuccess() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
	approval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")

	body := CreateWorkflowStepApprovalResponseRequest{
		ResponseType: app.WorkflowStepApprovalResponseTypeApprove,
		Note:         "lgtm",
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/%s/response", workflow.ID, step.ID, approval.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var result CreateWorkflowStepApprovalResponseResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	assert.NotEmpty(s.T(), result.ID)
	assert.Equal(s.T(), string(app.WorkflowStepApprovalResponseTypeApprove), result.Type)
	assert.Equal(s.T(), "lgtm", result.Note)

	// Verify response persisted in DB and linked to the approval.
	var dbResponse app.WorkflowStepApprovalResponse
	require.NoError(s.T(), s.deps.DB.Where("id = ?", result.ID).First(&dbResponse).Error)
	assert.Equal(s.T(), approval.ID, dbResponse.InstallWorkflowStepApprovalID)
	assert.Equal(s.T(), app.WorkflowStepApprovalResponseTypeApprove, dbResponse.Type)
}

func (s *InstallsServiceTestSuite) TestCreateApprovalResponseAlreadyExists() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
	approval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")

	// Create a response inline so the approval already has one.
	existingResponse := &app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approval.ID,
		Type:                          app.WorkflowStepApprovalResponseTypeApprove,
		Note:                          "already approved",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(existingResponse).Error)

	body := CreateWorkflowStepApprovalResponseRequest{
		ResponseType: app.WorkflowStepApprovalResponseTypeDeny,
		Note:         "too late",
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/%s/response", workflow.ID, step.ID, approval.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	// Handler returns ErrUser when approval already has a response.
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateApprovalResponseNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)

	body := CreateWorkflowStepApprovalResponseRequest{
		ResponseType: app.WorkflowStepApprovalResponseTypeApprove,
		Note:         "lgtm",
	}

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/nonexistent/response", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
