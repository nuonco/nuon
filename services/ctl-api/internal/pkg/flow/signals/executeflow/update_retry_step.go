package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

	switch step.Status.Status {
	case app.AwaitingApproval, app.Status("awaiting-approval"):
		// Step is awaiting approval - forward a retry response through the
		// approval mechanism. The step workflow will handle cloning and write
		// a retry directive.
		if _, err := workflowactivities.AwaitForwardApprovePlan(ctx, workflowactivities.ForwardApprovePlanRequest{
			StepID:       req.StepID,
			ResponseType: string(app.WorkflowStepApprovalResponseTypeRetryPlan),
		}); err != nil {
			return nil, fmt.Errorf("unable to forward retry approval for step %s: %w", req.StepID, err)
		}

		s.resumeRequested = true
		s.resumeRunType = app.WorkflowRunTypeRetry
		s.resumeStepID = req.StepID
		return &RetryStepResponse{
			WorkflowID: s.WorkflowID,
			Retryable:  true,
		}, nil

	case app.StatusError:
		if !step.Retryable {
			return &RetryStepResponse{
				WorkflowID: s.WorkflowID,
				Retryable:  false,
			}, nil
		}

		// Step failed during execution (not at approval). Create a clone step
		// directly via the step's create-step-retry update handler, then resume.
		if _, err := workflowactivities.AwaitForwardCreateStepRetry(ctx, workflowactivities.ForwardCreateStepRetryRequest{
			StepID: req.StepID,
		}); err != nil {
			return nil, fmt.Errorf("unable to create step retry for step %s: %w", req.StepID, err)
		}

		s.resumeRequested = true
		s.resumeRunType = app.WorkflowRunTypeRetry
		s.resumeStepID = req.StepID
		return &RetryStepResponse{
			WorkflowID: s.WorkflowID,
			Retryable:  true,
		}, nil

	default:
		return &RetryStepResponse{
			WorkflowID: s.WorkflowID,
			Retryable:  false,
		}, nil
	}
}
