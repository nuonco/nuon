package executeflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

const retryConductorParkWait = 30 * time.Second

const retryStepFlowOwnedVersion = "retry-step-flow-owned"

func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	defer s.beginUpdate()()

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if step.WorkflowStepGroupID == "" {
		return nil, fmt.Errorf("step %s has no group ID, cannot forward retry", req.StepID)
	}

	if s.Resident {
		return s.retryStepResident(ctx, req, step)
	}

	if workflow.GetVersion(ctx, retryStepFlowOwnedVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return s.retryStepLegacy(ctx, req, step)
	}

	if s.cancelRequested {
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow %s: %w", s.WorkflowID, err)
	}
	if flw.Status.Status == app.StatusCancelled {
		s.cancelRequested = true
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}

	flowOwned := s.awaitingResume || flw.Status.Status == app.StatusError
	if !flowOwned {
		if _, err := workflowactivities.AwaitForwardRetryStepToGroup(ctx, workflowactivities.ForwardRetryStepToGroupRequest{
			StepID:      req.StepID,
			StepGroupID: step.WorkflowStepGroupID,
		}); err != nil {
			return nil, fmt.Errorf("unable to forward retry to group: %w", err)
		}
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
	}

	// why: The conductor must be in its settled park before the step is mutated —
	// by then checkRetryable() has already run (so marking the step discarded
	// cannot flip the conductor into a terminal exit) and the group handler is
	// fully terminal (so no clone race with the group loop).
	if !s.awaitingResume {
		if !s.executeStarted {
			return nil, fmt.Errorf("workflow %s is no longer running and cannot be retried", s.WorkflowID)
		}
		woke, err := workflow.AwaitWithTimeout(ctx, retryConductorParkWait, func() bool {
			return s.awaitingResume || s.cancelRequested
		})
		if err != nil {
			return nil, fmt.Errorf("unable to wait for workflow %s to settle: %w", s.WorkflowID, err)
		}
		if !woke && !s.awaitingResume {
			return nil, fmt.Errorf("workflow %s is not awaiting a retry", s.WorkflowID)
		}
		if s.cancelRequested {
			return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
		}
	}

	if err := s.forwardRetryAndClone(ctx, req.StepID, step.GroupIdx); err != nil {
		return nil, err
	}

	s.markResumeRequested(ctx, app.WorkflowRunTypeRetry, req.StepID)

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}

func (s *Signal) retryStepResident(ctx workflow.Context, req RetryStepRequest, step *app.WorkflowStep) (*RetryStepResponse, error) {
	if s.retryInFlight == nil {
		s.retryInFlight = make(map[string]bool)
	}
	if s.retryInFlight[req.StepID] {
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
	}
	s.retryInFlight[req.StepID] = true
	defer delete(s.retryInFlight, req.StepID)

	stepDirective := directive.Step(step.ResultDirective)
	parked := stepDirective == directive.StepAwaitRetry || stepDirective == directive.StepAwaitApproval
	if !parked && step.Status.Status != app.StatusError {
		return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: false}, nil
	}

	if s.cancelRequested {
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow %s: %w", s.WorkflowID, err)
	}
	if flw.Status.Status == app.StatusCancelled {
		s.cancelRequested = true
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}

	if step.Retried {
		if _, err := workflowactivities.AwaitForwardCreateStepRetry(ctx, workflowactivities.ForwardCreateStepRetryRequest{
			StepID: req.StepID,
		}); err != nil {
			return nil, fmt.Errorf("unable to forward retry to step: %w", err)
		}
	} else if err := s.forwardRetryAndClone(ctx, req.StepID, step.GroupIdx); err != nil {
		return nil, err
	}

	s.markResumeRequested(ctx, app.WorkflowRunTypeRetry, req.StepID)

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}

func (s *Signal) forwardRetryAndClone(ctx workflow.Context, stepID string, groupIdx int) error {
	resp, err := workflowactivities.AwaitForwardCreateStepRetry(ctx, workflowactivities.ForwardCreateStepRetryRequest{
		StepID: stepID,
	})
	if err != nil {
		return fmt.Errorf("unable to forward retry to step: %w", err)
	}

	if directive.Step(resp.Directive) == directive.StepRetryGroup {
		if err := s.cloneGroupForRetry(ctx, groupIdx); err != nil {
			return fmt.Errorf("unable to clone group for retry: %w", err)
		}
		return nil
	}
	if err := executeworkflowstepgroup.CloneStepForRetry(ctx, stepID, s.WorkflowID); err != nil {
		return fmt.Errorf("unable to clone step for retry: %w", err)
	}
	return nil
}

// why: retryStepLegacy is the pre-retryStepFlowOwnedVersion command sequence, kept
// only so in-flight histories replay deterministically.
func (s *Signal) retryStepLegacy(ctx workflow.Context, req RetryStepRequest, step *app.WorkflowStep) (*RetryStepResponse, error) {
	_, err := workflowactivities.AwaitForwardRetryStepToGroup(ctx, workflowactivities.ForwardRetryStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to group: %w", err)
	}

	if s.awaitingResume || (s.Resident && !s.executeStarted) {
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

		s.markResumeRequested(ctx, app.WorkflowRunTypeRetry, req.StepID)
	}

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}
