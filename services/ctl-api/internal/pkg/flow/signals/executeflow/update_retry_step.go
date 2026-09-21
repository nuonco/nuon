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

// RetryStepRequest is the input for the "retry-step" update handler.
type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

// RetryStepResponse is the response from the "retry-step" update handler.
type RetryStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

// retryConductorParkWait bounds how long the retry update waits for the
// conductor to reach its settled park after a retryable error.
const retryConductorParkWait = 30 * time.Second

// retryStepFlowOwnedVersion gates the flow-owned retry path (flow lookup,
// park wait, direct create-step-retry forward) so an update handler that was
// in flight when the worker rolled keeps replaying the group-forwarded
// command sequence.
const retryStepFlowOwnedVersion = "retry-step-flow-owned"

// retryStepHandler retries an errored step.
//
// Two ownership cases:
//
//   - Live park (step parked in await-retry or awaiting-approval): the group's
//     sequential loop is still blocked on the step and owns the retry —
//     forward through the group; its loop reads the directive and clones.
//
//   - Failed run (flow status errored, or the conductor already parked at
//     awaiting-resume): the group handler already exited and nothing would
//     consume a directive written on the step — the flow owns the clone. It
//     waits for the conductor to reach its settled park (so checkRetryable
//     has already run and the group handler is fully terminal), forwards
//     create-step-retry directly to the step handler, clones per the
//     returned directive, and wakes the conductor.
//
// Forwarding directly to the step (skipping the group hop) also avoids
// update-with-start creating a zombie group handler run for a signal that has
// already finished.
//
// Flow: API → flow (here) → step → directive → flow clones + resumes
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	s.updatesInFlight++
	defer func() { s.updatesInFlight-- }()

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if step.WorkflowStepGroupID == "" {
		return nil, fmt.Errorf("step %s has no group ID, cannot forward retry", req.StepID)
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

	// The group handler stays blocked on the step for live parks (await-retry,
	// awaiting-approval), so it owns directive consumption and cloning — keep
	// the forwarded path. Only a failed run (errored status, or the conductor
	// already parked awaiting resume) leaves no live group loop.
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

	// The conductor must be in its settled park before the step is mutated —
	// by then checkRetryable() has already run (so marking the step discarded
	// cannot flip the conductor into a terminal exit) and the group handler is
	// fully terminal (so no clone race with the group loop).
	if !s.awaitingResume {
		// A fresh handler run on a terminal queue signal never re-drives the
		// conductor unless the host is resident, so there is nothing to wait for.
		if !s.Resident && !s.executeStarted {
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

	// Validate retryability, mark the step retried+discarded, and get the
	// directive that decides group- vs step-level cloning. The step handler
	// itself may have finished — update-with-start starts it just to serve
	// this update.
	resp, err := workflowactivities.AwaitForwardCreateStepRetry(ctx, workflowactivities.ForwardCreateStepRetryRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to step: %w", err)
	}

	if directive.Step(resp.Directive) == directive.StepRetryGroup {
		if err := s.cloneGroupForRetry(ctx, step.GroupIdx); err != nil {
			return nil, fmt.Errorf("unable to clone group for retry: %w", err)
		}
	} else if err := executeworkflowstepgroup.CloneStepForRetry(ctx, req.StepID, s.WorkflowID); err != nil {
		return nil, fmt.Errorf("unable to clone step for retry: %w", err)
	}

	s.markResumeRequested(ctx, app.WorkflowRunTypeRetry, req.StepID)

	return &RetryStepResponse{WorkflowID: s.WorkflowID, Retryable: true}, nil
}

// retryStepLegacy is the pre-retryStepFlowOwnedVersion command sequence, kept
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
