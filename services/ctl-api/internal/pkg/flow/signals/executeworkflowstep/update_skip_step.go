package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// SkipStepRequest is the input for the step-level "skip-step" update handler.
type SkipStepRequest struct{}

// SkipStepResponse is the response from the step-level "skip-step" update handler.
type SkipStepResponse struct{}

// skipStepHandler marks the step as user-skipped and writes a continue directive.
// This unblocks handleStepError's Await without going through Cancel, which
// would overwrite the skip status and directive.
func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
	// Mark step as user-skipped.
	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusUserSkipped,
			StatusHumanDescription: "Step was skipped by the user.",
			Metadata: map[string]any{
				"skipped": true,
			},
		},
	})

	// Write continue directive so the group proceeds.
	_ = setResultDirective(ctx, s.StepID, DirectiveContinue)

	// Unblock handleStepError's Await.
	s.skipped = true

	return &SkipStepResponse{}, nil
}
