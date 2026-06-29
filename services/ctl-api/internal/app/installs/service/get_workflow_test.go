package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowSuccess() {
	result := s.createTestInstallViaAPI()
	require.NotEmpty(s.T(), result.WorkflowID)

	path := fmt.Sprintf("/v1/workflows/%s", result.WorkflowID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var workflow app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &workflow))
	assert.Equal(s.T(), result.WorkflowID, workflow.ID)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/workflows/iwf_nonexistent_00000000", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowOmitsGroupStepsAndApprovalContents() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	group := s.deps.Seeder.CreateWorkflowStepGroup(s.ctx, s.T(), workflow.ID)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID, testseed.WithStepGroup(group.ID))
	s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "terraform plan output")

	path := fmt.Sprintf("/v1/workflows/%s", workflow.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var wf app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &wf))
	require.Len(s.T(), wf.Steps, 1)
	require.NotNil(s.T(), wf.Steps[0].Approval)
	require.Len(s.T(), wf.StepGroups, 1)
	assert.Equal(s.T(), group.ID, wf.StepGroups[0].ID)
	assert.Empty(s.T(), wf.StepGroups[0].Steps)

	// Contents is json:"-", so the omit can only be asserted against the
	// query result directly.
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	loaded, err := s.installsService.getWorkflow(c, s.testOrg.ID, workflow.ID)
	require.NoError(s.T(), err)
	require.Len(s.T(), loaded.Steps, 1)
	require.NotNil(s.T(), loaded.Steps[0].Approval)
	assert.Empty(s.T(), loaded.Steps[0].Approval.Contents)
	assert.Empty(s.T(), loaded.StepGroups[0].Steps)
}
