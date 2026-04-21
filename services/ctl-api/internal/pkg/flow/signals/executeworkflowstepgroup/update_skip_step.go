package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

// skipStepHandler marks the step as user-skipped. The sequential loop will
// skip over it on the next iteration since it's no longer pending.
func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	// Only errored or awaiting-approval steps can be skipped
	switch step.Status.Status {
	case app.StatusError, app.AwaitingApproval, app.Status("awaiting-approval"):
		// ok
	default:
		return &SkipStepResponse{Skippable: false}, nil
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

	// Wake up the step loop if it's waiting for user action.
	s.userActionReceived = true

	return &SkipStepResponse{Skippable: true}, nil
}
