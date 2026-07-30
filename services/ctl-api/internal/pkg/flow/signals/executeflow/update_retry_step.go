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

// retryStepHandler forwards the retry request through the group to the step.
// Legacy groups clone the retry themselves; resident flows clone here after
// descendants unwind so warm and cold updates share one durable owner.
//
// Flow: API → flow (here) → group → step → directive → clone owner
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	s.updatesInFlight++
	defer func() { s.updatesInFlight-- }()
	if s.retryInFlight == nil {
		s.retryInFlight = make(map[string]bool)
	}
	if s.retryInFlight[req.StepID] {
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
	}
	s.retryInFlight[req.StepID] = true
	defer delete(s.retryInFlight, req.StepID)

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if step.WorkflowStepGroupID == "" {
		return nil, fmt.Errorf("step %s has no group ID, cannot forward retry", req.StepID)
	}
	stepDirective := directive.Step(step.ResultDirective)
	residentManualRetry := s.Resident && (stepDirective == directive.StepAwaitRetry || stepDirective == directive.StepAwaitApproval)
	if s.Resident && !residentManualRetry {
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: false}, nil
	}
	wasRetried := step.Retried

	_, err = workflowactivities.AwaitForwardRetryStepToGroup(ctx, workflowactivities.ForwardRetryStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to group: %w", err)
	}

	// Resident flows own manual retry cloning after the child stack unwinds.
	// Legacy parked flows retain their existing cold clone path.
	if residentManualRetry || s.awaitingResume {
		updated, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
		if err != nil {
			return nil, fmt.Errorf("unable to re-read step %s: %w", req.StepID, err)
		}

		if !residentManualRetry || !wasRetried {
			if directive.Step(updated.ResultDirective) == directive.StepRetryGroup {
				if err := s.cloneGroupForRetry(ctx, updated.GroupIdx); err != nil {
					return nil, fmt.Errorf("unable to clone group for retry: %w", err)
				}
			} else if err := executeworkflowstepgroup.CloneStepForRetry(ctx, req.StepID, s.WorkflowID); err != nil {
				return nil, fmt.Errorf("unable to clone step for retry: %w", err)
			}
		}

		s.resumeRunType = app.WorkflowRunTypeRetry
		s.resumeStepID = req.StepID
		s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
		s.resumeRequested = true
	}

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}
