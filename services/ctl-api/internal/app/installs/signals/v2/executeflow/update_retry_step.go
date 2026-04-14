package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// RetryStepRequest is the input for the "retry-step" update handler.
type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

// RetryStepResponse is the response from the "retry-step" update handler.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	// Validate step is in a retryable state
	switch step.Status.Status {
	case app.StatusError:
		if !step.Retryable {
			return &RetryStepResponse{
				WorkflowID: s.InstallWorkflowID,
				Retryable:  false,
			}, nil
		}
	case app.AwaitingApproval, app.Status("awaiting-approval"):
		// Deny the approval so the step handler unblocks. The step handler's
		// RetryPlan response path handles cloning the step with incremented
		// GroupRetryIdx, so we don't need to set retryRequested here.
		if _, err := activities.AwaitDenyStepApproval(ctx, activities.DenyStepApprovalRequest{
			StepID: req.StepID,
		}); err != nil {
			return nil, fmt.Errorf("unable to deny approval for step %s: %w", req.StepID, err)
		}
		return &RetryStepResponse{
			WorkflowID: s.InstallWorkflowID,
			Retryable:  true,
		}, nil
	default:
		return &RetryStepResponse{
			WorkflowID: s.InstallWorkflowID,
			Retryable:  false,
		}, nil
	}

	s.retryRequested = true
	s.retryStepID = req.StepID
	s.retryOperation = "retry-step"
	return &RetryStepResponse{
		WorkflowID: s.InstallWorkflowID,
		Retryable:  true,
	}, nil
}
