package service

import (
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func (s *InstallsServiceTestSuite) TestRetryWorkflowRequiresMatchingInstallDeprecated() {
	owner := s.createTestInstall()
	other := s.createTestInstall()
	if other.ID > owner.ID {
		owner, other = other, owner
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: other.ID}).Update("name", owner.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: owner.ID}).Update("name", "prod").Error)
	siblingApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), siblingApp.ID)
	sibling := s.deps.Seeder.CreateInstall(s.ctx, s.T(), siblingApp)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: sibling.ID}).Update("name", "prod").Error)

	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), owner.ID, app.WorkflowTypeReprovision)
	otherWorkflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), other.ID, app.WorkflowTypeReprovision)

	for name, tc := range map[string]struct {
		installID  string
		workflowID string
		status     int
		retried    bool
	}{
		"wrong install":                        {other.ID, workflow.ID, http.StatusNotFound, false},
		"matching ID wins over name":           {owner.ID, workflow.ID, http.StatusCreated, true},
		"name collision cannot redirect retry": {owner.ID, otherWorkflow.ID, http.StatusNotFound, false},
		"ambiguous name":                       {"prod", workflow.ID, http.StatusBadRequest, false},
	} {
		s.Run(name, func() {
			step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), tc.workflowID,
				testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)), testseed.WithStepRetryable(true))
			body := RetryWorkflowRequest{WorkflowID: tc.workflowID, StepID: step.ID, Operation: RetryOperationRetryStep}
			rr := s.makeRequest(http.MethodPost, "/v1/installs/"+tc.installID+"/retry-workflow", body)
			require.Equal(s.T(), tc.status, rr.Code, "body: %s", rr.Body.String())
			var persisted app.WorkflowStep
			require.NoError(s.T(), s.deps.DB.Where(app.WorkflowStep{ID: step.ID}).First(&persisted).Error)
			require.Equal(s.T(), tc.retried, persisted.Retried)
			require.Equal(s.T(), tc.workflowID, persisted.InstallWorkflowID)
		})
	}
}
