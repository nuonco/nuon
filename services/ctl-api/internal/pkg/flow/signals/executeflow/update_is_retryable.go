package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// IsRetryableResponse is the response from the "is-retryable" update handler.
type IsRetryableResponse struct {
	Retryable bool   `json:"retryable"`
	StepID    string `json:"step_id"`
}

func (s *Signal) isRetryableHandler(ctx workflow.Context) (*IsRetryableResponse, error) {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return nil, err
	}

	// Find the latest errored step
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		if step.Status.Status == app.StatusError {
			return &IsRetryableResponse{
				Retryable: step.Retryable,
				StepID:    step.ID,
			}, nil
		}
	}

	return &IsRetryableResponse{Retryable: false}, nil
}
