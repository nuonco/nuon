package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
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

func (s *InstallsServiceTestSuite) TestCancelWorkflowPreservesData() {
	install := s.createTestInstall()
	for name, status := range map[string]app.Status{
		"pending":  app.StatusPending,
		"orphaned": app.StatusInProgress,
	} {
		for _, surface := range []string{"http", "deprecated", "mcp"} {
			s.Run(name+"/"+surface, func() {
				workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision,
					testseed.WithWorkflowStatus(app.NewCompositeStatus(s.ctx, status)),
					func(w *app.Workflow) {
						w.ApprovalOption = app.InstallApprovalOptionApproveAll
						w.PlanOnly = true
						w.Role = "custom-role"
						w.ResultDirective = "await-approval"
						w.Request = &app.WorkflowRequest{RequestID: name + "-" + surface, PinnedAppConfigID: s.testAppConfig.ID}
						w.StartedAt = time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
						w.Metadata = map[string]*string{"workflow-name-suffix": generics.ToPtr("acme"), "inputs_only": generics.ToPtr("true")}
					})
				var before app.Workflow
				require.NoError(s.T(), s.deps.DB.First(&before, "id = ?", workflow.ID).Error)

				if surface == "mcp" {
					_, _, err := s.installsService.mcpCancelWorkflow(s.workflowMCPContext(), nil, mcpCancelWorkflowInput{WorkflowID: workflow.ID})
					require.NoError(s.T(), err)
				} else {
					prefix := "/v1/workflows/"
					if surface == "deprecated" {
						prefix = "/v1/install-workflows/"
					}
					rr := s.makeRequest(http.MethodPost, prefix+workflow.ID+"/cancel", nil)
					require.Equal(s.T(), http.StatusAccepted, rr.Code, "body: %s", rr.Body.String())
				}

				var after app.Workflow
				require.NoError(s.T(), s.deps.DB.First(&after, "id = ?", workflow.ID).Error)
				require.Equal(s.T(), app.StatusCancelled, after.Status.Status)
				require.False(s.T(), after.FinishedAt.IsZero())
				require.Equal(s.T(), "Reprovisioned install (acme)", after.Name)
				require.Equal(s.T(), before.OwnerID, after.OwnerID)
				require.Equal(s.T(), before.OwnerType, after.OwnerType)
				require.Equal(s.T(), before.OrgID, after.OrgID)
				require.Equal(s.T(), before.Type, after.Type)
				require.Equal(s.T(), before.Metadata, after.Metadata)
				require.Equal(s.T(), before.ApprovalOption, after.ApprovalOption)
				require.Equal(s.T(), before.PlanOnly, after.PlanOnly)
				require.Equal(s.T(), before.Role, after.Role)
				require.Equal(s.T(), before.ResultDirective, after.ResultDirective)
				require.Equal(s.T(), before.Request, after.Request)
				require.True(s.T(), before.StartedAt.Equal(after.StartedAt))
			})
		}
	}
}
