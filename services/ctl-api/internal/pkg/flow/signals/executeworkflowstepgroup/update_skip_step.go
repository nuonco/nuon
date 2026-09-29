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

type SkipStepRequest struct {
	StepID string `json:"step_id"`
}

type SkipStepResponse struct {
	Skippable bool   `json:"skippable"`
	Directive string `json:"directive,omitempty"`
}

func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if !step.Skippable {
		return &SkipStepResponse{Skippable: false}, nil
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

	activities.AwaitForwardSkipStep(ctx, activities.ForwardSkipStepRequest{
		StepID: req.StepID,
	})

	resp := &SkipStepResponse{Skippable: true}
	if s.ResidentFlow {
		resp.Directive = string(skipDirective)
	}
	return resp, nil
}
