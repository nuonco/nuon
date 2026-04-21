package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// RetryStepRequest is the input for the "retry-step" group update handler.
type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

// RetryStepResponse is the response from the "retry-step" group update handler.
type RetryStepResponse struct {
	Retryable bool   `json:"retryable"`
	Directive string `json:"directive"`
}

// retryStepHandler is called when the group signal is alive (e.g. during
// awaitUserAction after a step failure). It tells the step it was retried,
// clones it, and wakes the sequential loop to pick up the clone.
//
// When the group is dead, the flow's retryStepHandler handles it instead
// by cloning and re-dispatching a new group signal.
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	// Tell the step it was retried — marks as discarded, returns directive.
	resp, err := activities.AwaitForwardCreateStepRetry(ctx, activities.ForwardCreateStepRetryRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to step %s: %w", req.StepID, err)
	}

	switch resp.Directive {
	case DirectiveRetryGroup:
		if err := s.writeWorkflowDirective(ctx, DirectiveRetryGroup); err != nil {
			return nil, fmt.Errorf("unable to write retry-group directive: %w", err)
		}
		s.retryGroupRequested = true

	default:
		// Clone the step. The sequential loop (in awaitUserAction) will
		// pick up the clone on its next iteration.
		if err := cloneStepForRetry(ctx, req.StepID, s.WorkflowID); err != nil {
			return nil, fmt.Errorf("unable to clone step for retry: %w", err)
		}
	}

	s.userActionReceived = true

	return &RetryStepResponse{
		Retryable: true,
		Directive: resp.Directive,
	}, nil
}
