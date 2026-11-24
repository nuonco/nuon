package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestRetryWorkflowRequiresMatchingInstallDeprecated() {
	owner := s.createTestInstall()
	other := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), owner.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID,
		testseed.WithStepStatus(app.NewCompositeStatus(s.ctx, app.StatusError)), testseed.WithStepRetryable(true))
	body := RetryWorkflowRequest{WorkflowID: workflow.ID, StepID: step.ID, Operation: RetryOperationRetryStep}

	for _, tc := range []struct {
		name      string
		installID string
		status    int
		retried   bool
	}{
		{"wrong install", other.ID, http.StatusNotFound, false},
		{"matching install", owner.ID, http.StatusCreated, true},
	} {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, "/v1/installs/"+tc.installID+"/retry-workflow", body)
			require.Equal(s.T(), tc.status, rr.Code, "body: %s", rr.Body.String())
			var persisted app.WorkflowStep
			require.NoError(s.T(), s.deps.DB.First(&persisted, "id = ?", step.ID).Error)
			require.Equal(s.T(), tc.retried, persisted.Retried)
			require.Equal(s.T(), workflow.ID, persisted.InstallWorkflowID)
		})
	}
}
