package executeflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

// skipStepFlowOwnedVersion gates the flow-owned skip path (flow lookup, park
// wait, direct status update) so an update handler that was in flight when
// the worker rolled keeps replaying the group-forwarded command sequence.
const skipStepFlowOwnedVersion = "skip-step-flow-owned"

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
	defer s.beginUpdate()()

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if s.Resident {
		return s.skipStepResident(ctx, req, step)
	}

	if workflow.GetVersion(ctx, skipStepFlowOwnedVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return s.skipStepLegacy(ctx, req, step)
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
		// A fresh handler run on a terminal queue signal never re-drives the
		// conductor unless the host is resident, so there is nothing to wait for.
		if !s.Resident && !s.executeStarted {
			return nil, fmt.Errorf("workflow %s is no longer running and cannot be skipped", s.WorkflowID)
		}
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

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: req.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to mark step %s as skipped: %w", req.StepID, err)
	}

	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if sg, ok := step.QueueSignal.Signal.(signal.SignalWithSkipGroup); ok && sg.SkipGroup() {
			s.discardRemainingGroupSteps(ctx, step)
		}
	}

	s.markResumeRequested(ctx, app.WorkflowRunTypeSkip, req.StepID)

	// Skippable is unconditionally true here: the flow owns this path, so
	// there is no group response to relay. Non-skippable steps already
	// returned false above, and the step is now marked user-skipped with the
	// conductor woken past it.
	return &SkipStepResponse{
		WorkflowID: s.WorkflowID,
		Skippable:  true,
	}, nil
}

// discardRemainingGroupSteps is the flow-owned counterpart of the group's
// StepSkipGroup handling. The group loop that would have consumed that
// directive has already exited on a failed run, so the steps after the skipped
// one are discarded here and the resumed group finds nothing left to run.
func (s *Signal) discardRemainingGroupSteps(ctx workflow.Context, skipped *app.WorkflowStep) {
	l, _ := log.WorkflowLogger(ctx)

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		if l != nil {
			l.Warn("skip-group: unable to get flow steps", zap.String("step_id", skipped.ID), zap.Error(err))
		}
		return
	}

	for _, st := range steps {
		sameGroup := st.GroupIdx == skipped.GroupIdx
		if skipped.WorkflowStepGroupID != "" {
			sameGroup = st.WorkflowStepGroupID == skipped.WorkflowStepGroupID
		}
		if !sameGroup || st.Idx <= skipped.Idx || isStepTerminal(st.Status.Status) {
			continue
		}
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: st.ID,
			Status: app.CompositeStatus{
				Status: app.StatusDiscarded,
				Metadata: map[string]any{
					"reason": fmt.Sprintf("group step %s triggered stop", skipped.ID),
				},
			},
		}); err != nil && l != nil {
			l.Warn("skip-group: failed to discard remaining step", zap.String("step_id", st.ID), zap.Error(err))
		}
	}
}

// skipStepResident handles skip on a resident host. A step parked in
// await-retry or await-approval has no live group loop, so the flow marks it
// skipped, writes the step and group directives the resident conductor reads,
// and seeds the resume. Any other state still forwards through the group.
func (s *Signal) skipStepResident(ctx workflow.Context, req SkipStepRequest, step *app.WorkflowStep) (*SkipStepResponse, error) {
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

	if !step.Skippable {
		return &SkipStepResponse{WorkflowID: s.WorkflowID, Skippable: false}, nil
	}

	stepDirective := directive.Step(step.ResultDirective)
	residentParked := stepDirective == directive.StepAwaitRetry || stepDirective == directive.StepAwaitApproval

	resp, err := workflowactivities.AwaitForwardSkipStepToGroup(ctx, workflowactivities.ForwardSkipStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil && !residentParked {
		return nil, fmt.Errorf("unable to forward skip-step to group: %w", err)
	}
	if err != nil {
		if l, _ := log.WorkflowLogger(ctx); l != nil {
			l.Warn("skip-step: unable to forward skip to unwound group",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
	}

	if residentParked {
		skipDirective := residentSkipDirective(step)
		if resp != nil && resp.Directive != "" {
			skipDirective = directive.Step(resp.Directive)
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusUserSkipped,
				StatusHumanDescription: "Step was skipped by the user.",
				Metadata: map[string]any{
					"skipped": true,
				},
			},
		}); err != nil {
			return nil, fmt.Errorf("unable to mark step %s as skipped: %w", step.ID, err)
		}
		if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, workflowactivities.UpdateFlowStepResultDirectiveRequest{
			StepID:    step.ID,
			Directive: string(skipDirective),
		}); err != nil {
			return nil, fmt.Errorf("unable to write skip directive: %w", err)
		}
		if err := s.repairResidentSkippedGroup(ctx, step, skipDirective); err != nil {
			return nil, err
		}

		s.markResumeRequested(ctx, app.WorkflowRunTypeSkip, req.StepID)
	}

	skippable := true
	if resp != nil {
		skippable = resp.Skippable
	}

	// Set unconditionally (like approve/unpause): gating on awaitingResume
	// loses skips that land between the group unwinding and the flow parking.
	// A skip that lands mid-run is dropped by the main loop's stale-resume
	// reset.
	if !residentParked && skippable {
		s.markResumeRequested(ctx, app.WorkflowRunTypeSkip, req.StepID)
	}

	return &SkipStepResponse{
		WorkflowID: s.WorkflowID,
		Skippable:  skippable,
	}, nil
}

func residentSkipDirective(step *app.WorkflowStep) directive.Step {
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if skipGroup, ok := step.QueueSignal.Signal.(signal.SignalWithSkipGroup); ok && skipGroup.SkipGroup() {
			return directive.StepSkipGroup
		}
	}
	return directive.StepContinue
}

func (s *Signal) repairResidentSkippedGroup(ctx workflow.Context, step *app.WorkflowStep, skipDirective directive.Step) error {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return fmt.Errorf("unable to load group steps after skip: %w", err)
	}

	hasPending := false
	for _, candidate := range steps {
		sameGroup := candidate.GroupIdx == step.GroupIdx
		if step.WorkflowStepGroupID != "" {
			sameGroup = candidate.WorkflowStepGroupID == step.WorkflowStepGroupID
		}
		if !sameGroup || candidate.ID == step.ID || isStepTerminal(candidate.Status.Status) {
			continue
		}

		if skipDirective == directive.StepSkipGroup && candidate.Idx > step.Idx {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: candidate.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusDiscarded,
					StatusHumanDescription: "Step was discarded after the group was skipped.",
				},
			}); err != nil {
				return fmt.Errorf("unable to discard skipped group step %s: %w", candidate.ID, err)
			}
			continue
		}
		hasPending = true
	}

	if step.WorkflowStepGroupID == "" {
		return nil
	}

	groupDirective := directive.GroupContinue
	if skipDirective == directive.StepSkipGroup {
		groupDirective = directive.GroupSkipGroup
	}
	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepGroupResultDirective(ctx, workflowactivities.UpdateFlowStepGroupResultDirectiveRequest{
		StepGroupID: step.WorkflowStepGroupID,
		Directive:   string(groupDirective),
	}); err != nil {
		return fmt.Errorf("unable to update skipped group directive: %w", err)
	}

	groupStatus := app.StatusSuccess
	groupDescription := "group completed after step was skipped"
	if hasPending {
		groupStatus = app.StatusPending
		groupDescription = "group pending after step was skipped"
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.WorkflowStepGroupID,
		Status: app.CompositeStatus{
			Status:                 groupStatus,
			StatusHumanDescription: groupDescription,
		},
	}); err != nil {
		return fmt.Errorf("unable to update skipped group status: %w", err)
	}

	return nil
}

// skipStepLegacy is the pre-skipStepFlowOwnedVersion command sequence, kept
// only so in-flight histories replay deterministically.
func (s *Signal) skipStepLegacy(ctx workflow.Context, req SkipStepRequest, step *app.WorkflowStep) (*SkipStepResponse, error) {
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
