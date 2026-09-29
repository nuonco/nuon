package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type SkipStepRequest struct{}

type SkipStepResponse struct{}

func (s *Signal) skipStepHandler(ctx workflow.Context, req SkipStepRequest) (*SkipStepResponse, error) {
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

	_ = setResultDirective(ctx, s.StepID, DirectiveContinue)

	s.skipped = true

	return &SkipStepResponse{}, nil
}
