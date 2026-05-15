package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
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
	Skippable bool `json:"skippable"`
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
	skipDirective := directive.StepContinue
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if sg, ok := step.QueueSignal.Signal.(signal.SignalWithSkipGroup); ok && sg.SkipGroup() {
			skipDirective = directive.StepSkipGroup
		}
	}

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

	return &SkipStepResponse{Skippable: true}, nil
}
