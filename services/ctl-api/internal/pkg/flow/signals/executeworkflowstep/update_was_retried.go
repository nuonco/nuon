package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// WasRetriedResponse is the response from the "was-retried" update handler.
type WasRetriedResponse struct{}

// wasRetriedHandler is called by the step-group after it has decided to retry
// this step. It marks the step as discarded so the step signal stops waiting
// (e.g. for approval) and sets s.finished so the step-finished handler unblocks.
//
// The actual cloning and re-dispatch happens in the step-group, not here.
func (s *Signal) wasRetriedHandler(ctx workflow.Context) (*WasRetriedResponse, error) {
	_ = activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
		StepID: s.StepID,
	})

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusDiscarded,
			StatusHumanDescription: "Step was retried.",
			Metadata: map[string]any{
				"retry_type": "was-retried",
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to mark step as discarded")
	}

	if err := setResultDirective(ctx, s.StepID, DirectiveRetry); err != nil {
		return nil, errors.Wrap(err, "unable to set result directive")
	}

	// Unblock waitForApprovalResponse if the step is waiting for approval.
	s.retried = true
	return &WasRetriedResponse{}, nil
}
