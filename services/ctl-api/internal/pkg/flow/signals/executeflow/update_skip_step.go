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

type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

const skipConductorParkWait = 30 * time.Second

const skipStepFlowOwnedVersion = "skip-step-flow-owned"

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

	// why: The conductor must be in its settled park before the step is mutated —
	// by then checkRetryable() has already read the errored step, so flipping
	// it to user-skipped cannot race the park-vs-terminal decision.
	if !s.awaitingResume {
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

	return &SkipStepResponse{
		WorkflowID: s.WorkflowID,
		Skippable:  true,
	}, nil
}

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

// why: skipStepLegacy is the pre-skipStepFlowOwnedVersion command sequence, kept
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
