package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// SkipStepRequest is the input for the "skip-step" group update handler.
type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

// SkipStepResponse is the response from the "skip-step" group update handler.
type SkipStepResponse struct {
	Skippable bool   `json:"skippable"`
	Directive string `json:"directive,omitempty"`
}

// skipStepHandler marks the step as user-skipped, writes a continue directive,
// and cancels the step signal so its Execute() unblocks. The group's sequential
// loop then reads the directive and proceeds.
func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if !step.Skippable {
		return &SkipStepResponse{Skippable: false}, nil
	}

	// Mark step as user-skipped.
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: req.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to mark step %s as skipped: %w", req.StepID, err)
	}

	// Determine the directive: skip-group if the signal declares it, otherwise continue.
	skipDirective := skipStepDirective(step)

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID:    req.StepID,
		Directive: string(skipDirective),
	}); err != nil {
		return nil, fmt.Errorf("unable to write skip directive: %w", err)
	}

	// Send skip-step to the step signal to unblock its Execute() cleanly
	// without going through Cancel (which would overwrite the skip status).
	activities.AwaitForwardSkipStep(ctx, activities.ForwardSkipStepRequest{
		StepID: req.StepID,
	})

	resp := &SkipStepResponse{Skippable: true}
	if s.ResidentFlow {
		resp.Directive = string(skipDirective)
	}
	return resp, nil
}

func SkipSettledStep(ctx workflow.Context, step *app.WorkflowStep) error {
	s := &Signal{WorkflowID: step.InstallWorkflowID, StepGroupID: step.WorkflowStepGroupID, GroupIdx: step.GroupIdx}
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	steps, err := s.getGroupSteps(ctx)
	if err != nil {
		return err
	}
	skipDirective := skipStepDirective(step)
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID: step.ID, Directive: string(skipDirective),
	}); err != nil {
		return err
	}
	if skipDirective == directive.StepSkipGroup {
		if err := s.skipGroup(ctx, l, steps, step.ID, app.StatusDiscarded, app.CompositeStatus{}); err != nil {
			return err
		}
	} else if err := s.writeStepGroupDirective(ctx, directive.GroupContinue); err != nil {
		return err
	}

	groupStatus := app.StatusSuccess
	for _, candidate := range steps {
		if candidate.ID == step.ID || isTerminalStatus(candidate.Status.Status) {
			continue
		}
		if skipDirective == directive.StepSkipGroup && candidate.Idx > step.Idx {
			continue
		}
		groupStatus = app.StatusPending
		break
	}
	if s.StepGroupID != "" {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.StepGroupID, Status: app.CompositeStatus{Status: groupStatus},
		}); err != nil {
			return err
		}
	}
	return statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
			Metadata:               map[string]any{"skipped": true},
		},
	})
}

func skipStepDirective(step *app.WorkflowStep) directive.Step {
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if sg, ok := step.QueueSignal.Signal.(signal.SignalWithSkipGroup); ok && sg.SkipGroup() {
			return directive.StepSkipGroup
		}
	}
	return directive.StepContinue
}
