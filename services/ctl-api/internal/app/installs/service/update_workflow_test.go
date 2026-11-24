package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
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

func (s *InstallsServiceTestSuite) TestApproveAllOnlyAnswersWaitingApprovals() {
	install := s.createTestInstall()
	for _, surface := range []string{"http", "mcp"} {
		s.Run(surface, func() {
			workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
			other := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
			approvals := make(map[string]*app.WorkflowStepApproval)
			for name, tc := range map[string]struct {
				workflowID string
				status     app.Status
			}{
				"waiting":        {workflow.ID, app.AwaitingApproval},
				"waiting legacy": {workflow.ID, app.Status("awaiting-approval")},
				"answered":       {workflow.ID, app.AwaitingApproval},
				"not waiting":    {workflow.ID, app.StatusPending},
				"other workflow": {other.ID, app.AwaitingApproval},
			} {
				step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), tc.workflowID,
					testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, tc.status)),
					func(step *app.WorkflowStep) { step.ExecutionType = app.WorkflowStepExecutionTypeApproval })
				approvals[name] = s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "")
			}
			answered := &app.WorkflowStepApprovalResponse{
				InstallWorkflowStepApprovalID: approvals["answered"].ID,
				Type:                          app.WorkflowStepApprovalResponseTypeDeny,
				Note:                          "keep this decision",
			}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(answered).Error)

			if surface == "mcp" {
				_, _, err := s.installsService.mcpApproveAll(s.workflowMCPContext(), nil, mcpApproveAllInput{WorkflowID: workflow.ID})
				require.NoError(s.T(), err)
			} else {
				option := app.InstallApprovalOptionApproveAll
				rr := s.makeRequest(http.MethodPatch, "/v1/workflows/"+workflow.ID, UpdateWorkflowRequest{ApprovalOption: &option})
				require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())
			}

			for name, approval := range approvals {
				var responses []app.WorkflowStepApprovalResponse
				require.NoError(s.T(), s.deps.DB.Where(app.WorkflowStepApprovalResponse{InstallWorkflowStepApprovalID: approval.ID}).Find(&responses).Error)
				switch name {
				case "waiting", "waiting legacy":
					require.Len(s.T(), responses, 1)
					require.Equal(s.T(), app.WorkflowStepApprovalResponseTypeApprove, responses[0].Type)
				case "answered":
					require.Len(s.T(), responses, 1)
					require.Equal(s.T(), answered.ID, responses[0].ID)
					require.Equal(s.T(), answered.Type, responses[0].Type)
					require.Equal(s.T(), answered.Note, responses[0].Note)
				default:
					require.Empty(s.T(), responses, name)
				}
			}
			var updated app.Workflow
			require.NoError(s.T(), s.deps.DB.First(&updated, "id = ?", workflow.ID).Error)
			require.Equal(s.T(), app.InstallApprovalOptionApproveAll, updated.ApprovalOption)
			var untouched app.Workflow
			require.NoError(s.T(), s.deps.DB.First(&untouched, "id = ?", other.ID).Error)
			require.Equal(s.T(), app.InstallApprovalOptionPrompt, untouched.ApprovalOption)
		})
	}
}

func (s *InstallsServiceTestSuite) TestApproveAllRejectsOtherOrg() {
	foreignCtx, _ := s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	foreignApp := s.deps.Seeder.CreateApp(foreignCtx, s.T())
	workflow := s.deps.Seeder.CreateWorkflow(foreignCtx, s.T(), foreignApp.ID, app.WorkflowTypeAppConfigBuild,
		testseed.WithWorkflowOwnerType("apps"))
	step := s.deps.Seeder.CreateWorkflowStep(foreignCtx, s.T(), workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(foreignCtx, app.AwaitingApproval)),
		func(step *app.WorkflowStep) { step.ExecutionType = app.WorkflowStepExecutionTypeApproval })
	approval := s.deps.Seeder.CreateWorkflowStepApproval(foreignCtx, s.T(), step.ID, app.TerraformPlanApprovalType, "")

	option := app.InstallApprovalOptionApproveAll
	rr := s.makeRequest(http.MethodPatch, "/v1/workflows/"+workflow.ID, UpdateWorkflowRequest{ApprovalOption: &option})
	require.Equal(s.T(), http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())
	_, _, err := s.installsService.mcpApproveAll(s.workflowMCPContext(), nil, mcpApproveAllInput{WorkflowID: workflow.ID})
	require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)

	var unchanged app.Workflow
	require.NoError(s.T(), s.deps.DB.First(&unchanged, "id = ?", workflow.ID).Error)
	require.Equal(s.T(), app.InstallApprovalOptionPrompt, unchanged.ApprovalOption)
	var count int64
	require.NoError(s.T(), s.deps.DB.Model(&app.WorkflowStepApprovalResponse{}).
		Where(app.WorkflowStepApprovalResponse{InstallWorkflowStepApprovalID: approval.ID}).Count(&count).Error)
	require.Zero(s.T(), count)
}
