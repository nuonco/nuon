package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type RunnerJobCompletedRequest struct {
	JobID string `json:"job_id"`
}

func (s *Signal) runnerJobCompleted(ctx workflow.Context, req RunnerJobCompletedRequest) error {
	defer s.beginUpdate()()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	var stepID string
	if err := workflow.ExecuteActivity(ctx, (*workflowactivities.Activities).PrepareRunnerJobResume, s.WorkflowID, req.JobID).Get(ctx, &stepID); err != nil {
		return err
	}
	if stepID != "" {
		s.markResumeRequested(ctx, app.WorkflowRunTypeResume, stepID)
	}
	return nil
}
