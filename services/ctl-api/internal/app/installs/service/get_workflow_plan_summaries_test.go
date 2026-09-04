package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowPlanSummaries() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	terraformStep := s.deps.Seeder.CreateWorkflowStep(
		s.ctx,
		s.T(),
		workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.AwaitingApproval)),
	)
	terraformStep.Name = "sync and plan api"
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: terraformStep.ID}).
		Update("name", terraformStep.Name).Error)
	terraformApproval := s.deps.Seeder.CreateWorkflowStepApproval(
		s.ctx,
		s.T(),
		terraformStep.ID,
		app.TerraformPlanApprovalType,
		`{"resource_changes":[{"change":{"actions":["create"]}},{"change":{"actions":["update"]}}]}`,
	)

	installStep := s.deps.Seeder.CreateWorkflowStep(
		s.ctx,
		s.T(),
		workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.WorkflowStepApprovalStatusApproved)),
	)
	installStep.Name = "approve install creation"
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: installStep.ID}).
		Update("name", installStep.Name).Error)
	installApproval := s.deps.Seeder.CreateWorkflowStepApproval(
		s.ctx,
		s.T(),
		installStep.ID,
		app.InstallCreationApprovalType,
		`{"installs":[]}`,
	)

	noopStep := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
	s.deps.Seeder.CreateWorkflowStepApproval(
		s.ctx,
		s.T(),
		noopStep.ID,
		app.NoopApprovalType,
		`{}`,
	)

	path := fmt.Sprintf("/v1/workflows/%s/plan-summaries", workflow.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var summaries []app.StepChangeSummary
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &summaries))
	require.Len(s.T(), summaries, 2)

	require.Equal(s.T(), terraformStep.ID, summaries[0].StepID)
	require.Equal(s.T(), terraformApproval.ID, summaries[0].ApprovalID)
	require.Equal(s.T(), app.StepChangeStatusPendingApproval, summaries[0].Status)
	require.Equal(s.T(), app.StepChangeCounts{Create: 1, Update: 1}, summaries[0].Counts)
	require.True(s.T(), summaries[0].HasDetail)

	require.Equal(s.T(), installStep.ID, summaries[1].StepID)
	require.Equal(s.T(), installApproval.ID, summaries[1].ApprovalID)
	require.Equal(s.T(), app.StepChangeStatusApproved, summaries[1].Status)
	require.Empty(s.T(), summaries[1].Counts)
	require.False(s.T(), summaries[1].HasDetail)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowPlanSummariesNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/workflows/iwf_nonexistent_00000000/plan-summaries", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
