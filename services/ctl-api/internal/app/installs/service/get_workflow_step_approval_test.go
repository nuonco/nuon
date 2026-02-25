package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowStepApprovalSuccess() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
	approval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/%s", workflow.ID, step.ID, approval.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var result app.WorkflowStepApproval
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Equal(s.T(), approval.ID, result.ID)
	assert.Equal(s.T(), app.TerraformPlanApprovalType, result.Type)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowStepApprovalNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/nonexistent", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowStepApprovalStepNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	path := fmt.Sprintf("/v1/workflows/%s/steps/nonexistent/approvals/nonexistent", workflow.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
