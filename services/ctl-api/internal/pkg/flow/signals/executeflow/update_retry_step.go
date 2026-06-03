package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
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

// retryStepHandler forwards the retry request to the group, which forwards to
// the step. The step's createStepRetryHandler writes the terminal directive and
// unblocks Execute(). The queue signal completes, the group reads the directive,
// and the group's sequential loop handles cloning.
//
// Flow: API → flow (here) → group → step → directive → group clones
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if step.WorkflowStepGroupID == "" {
		return nil, fmt.Errorf("step %s has no group ID, cannot forward retry", req.StepID)
	}

	_, err = workflowactivities.AwaitForwardRetryStepToGroup(ctx, workflowactivities.ForwardRetryStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to group: %w", err)
	}

	// Parked flow: the group isn't live to clone, so clone here and wake the loop.
	if s.awaitingResume {
		updated, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
		if err != nil {
			return nil, fmt.Errorf("unable to re-read step %s: %w", req.StepID, err)
		}

		if directive.Step(updated.ResultDirective) == directive.StepRetryGroup {
			if err := s.cloneGroupForRetry(ctx, updated.GroupIdx); err != nil {
				return nil, fmt.Errorf("unable to clone group for retry: %w", err)
			}
		} else if err := executeworkflowstepgroup.CloneStepForRetry(ctx, req.StepID, s.WorkflowID); err != nil {
			return nil, fmt.Errorf("unable to clone step for retry: %w", err)
		}

		s.resumeRequested = true
		s.resumeRunType = app.WorkflowRunTypeRetry
		s.resumeStepID = req.StepID
		s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
	}

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}
