package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// IsRetryableResponse is the response from the "is-retryable" update handler.
type IsRetryableResponse struct {
	Retryable bool   `json:"retryable"`
	Skippable bool   `json:"skippable"`
	StepID    string `json:"step_id"`
}

func (s *Signal) isRetryableHandler(ctx workflow.Context) (*IsRetryableResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, err
	}
	return &IsRetryableResponse{
		Retryable: step.Retryable,
		Skippable: step.Skippable,
		StepID:    step.ID,
	}, nil
}
