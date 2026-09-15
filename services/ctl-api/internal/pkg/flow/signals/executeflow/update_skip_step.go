package executeflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// SkipStepRequest is the input for the "skip-step" update handler.
type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

// SkipStepResponse is the response from the "skip-step" update handler.
type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

// skipConductorParkWait bounds how long the skip update waits for the
// conductor to reach its settled park after a terminal error.
const skipConductorParkWait = 30 * time.Second

// skipParkWaitVersion gates the wait-for-parked-conductor behavior so
// in-flight workflows replaying pre-change history keep the old command
// sequence (the wait issues a timer command).
const skipParkWaitVersion = "skip-step-park-wait"

// skipStepHandler skips an errored step.
//
// Two ownership cases (mirrors retryStepHandler):
//
//   - Live park (step parked in await-retry or awaiting-approval): the group's
//     sequential loop is still blocked on the step and owns the skip —
//     forward through the group; its loop reads the directive and proceeds.
//
//   - Failed run (flow status errored, or the conductor already parked at
//     awaiting-resume): the group handler already exited and nothing would
//     consume a directive written on the step — the flow owns the skip. It
//     waits for the conductor to reach its settled park, marks the step
//     user-skipped (no clone), and wakes the conductor at the skipped step's
//     group; the resumed run treats the skipped step as terminal and advances
//     past it.
//
// Flow: API → flow (here) → mark skipped → conductor resumes past it
func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	s.updatesInFlight++
	defer func() { s.updatesInFlight-- }()

	// Look up the step to get its group ID for direct group lookup.
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if s.cancelRequested {
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow %s: %w", s.WorkflowID, err)
	}
	if flw.Status.Status == app.StatusCancelled {
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}

	// The group handler stays blocked on the step for live parks (await-retry,
	// awaiting-approval), so it owns directive consumption — keep the forwarded
	// path. Only a failed run (errored status, or the conductor already parked
	// awaiting resume) leaves no live group loop.
	flowOwned := s.awaitingResume || flw.Status.Status == app.StatusError
	if !flowOwned {
		resp, err := workflowactivities.AwaitForwardSkipStepToGroup(ctx, workflowactivities.ForwardSkipStepToGroupRequest{
			StepID:      req.StepID,
			StepGroupID: step.WorkflowStepGroupID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to forward skip-step to group: %w", err)
		}
		return &SkipStepResponse{
			WorkflowID: s.WorkflowID,
			Skippable:  resp.Skippable,
		}, nil
	}

	if step.Status.Status != app.StatusError || !step.Skippable {
		return &SkipStepResponse{
			WorkflowID: s.WorkflowID,
			Skippable:  false,
		}, nil
	}

	// The conductor must be in its settled park before the step is mutated —
	// by then checkRetryable() has already read the errored step, so flipping
	// it to user-skipped cannot race the park-vs-terminal decision.
	if !s.awaitingResume {
		if workflow.GetVersion(ctx, skipParkWaitVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			woke, err := workflow.AwaitWithTimeout(ctx, skipConductorParkWait, func() bool {
				return s.awaitingResume || s.cancelRequested
			})
			if err != nil {
				return nil, fmt.Errorf("unable to wait for workflow %s to settle: %w", s.WorkflowID, err)
			}
			if !woke && !s.awaitingResume {
				return nil, fmt.Errorf("workflow %s is not awaiting a skip", s.WorkflowID)
			}
			if s.cancelRequested {
				return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
			}
		}
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: req.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to mark step %s as skipped: %w", req.StepID, err)
	}

	// Publish the wake state only after resumeStartIdx is fully computed:
	// findGroupPositionForStep yields on activities, and the conductor wakes
	// on resumeRequested, so the flag must go out last.
	s.resumeRunType = app.WorkflowRunTypeSkip
	s.resumeStepID = req.StepID
	s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
	s.resumeRequested = true

	// Skippable is unconditionally true here: the flow owns this path, so
	// there is no group response to relay. Non-skippable steps already
	// returned false above, and the step is now marked user-skipped with the
	// conductor woken past it.
	return &SkipStepResponse{
		WorkflowID: s.WorkflowID,
		Skippable:  true,
	}, nil
}
