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
// Resident hosts always own the clone: the group handler returns on
// await-retry instead of blocking, so nothing downstream would consume a
// directive written on the step. retryStepResident forwards create-step-retry
// directly to the step, clones per the returned directive, and seeds the
// resume so both a warm loop and a cold re-warm pick it up.
//
// Legacy hosts have two ownership cases:
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
		// conductor, so there is nothing to wait for.
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

// retryStepResident handles manual retry on a resident host. The step must be
// parked (await-retry or awaiting-approval) or terminally errored — the same
// states legacy accepted; the step handler still enforces Retryable and the
// retry ceiling. A step that was already retried (duplicate request after the
// clone landed) only re-seeds the resume so the loop wakes without cloning
// twice.
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

// forwardRetryAndClone asks the step handler to validate retryability, mark
// the step retried+discarded, and return the directive that decides group- vs
// step-level cloning, then clones accordingly. The step handler itself may
// have finished — update-with-start starts it just to serve this update.
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
