package executeflow

import (
	"fmt"

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

func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	s.updatesInFlight++
	defer func() { s.updatesInFlight-- }()

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}
	if s.Resident && !step.Skippable {
		return &SkipStepResponse{WorkflowID: s.WorkflowID, Skippable: false}, nil
	}

	residentAwaitRetry := s.Resident && directive.Step(step.ResultDirective) == directive.StepAwaitRetry

	resp, err := workflowactivities.AwaitForwardSkipStepToGroup(ctx, workflowactivities.ForwardSkipStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil && !residentAwaitRetry {
		return nil, fmt.Errorf("unable to forward skip-step to group: %w", err)
	}
	if err != nil {
		if l, _ := log.WorkflowLogger(ctx); l != nil {
			l.Warn("skip-step: unable to forward skip to unwound group",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
	}

	if residentAwaitRetry {
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

		s.resumeRunType = app.WorkflowRunTypeSkip
		s.resumeStepID = req.StepID
		s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
		s.resumeRequested = true
	}

	skippable := true
	if resp != nil {
		skippable = resp.Skippable
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
