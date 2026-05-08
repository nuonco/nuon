package executeworkflowstep

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// awaitAndHandleApproval writes the await-approval directive, waits for a
// user response, and dispatches to the appropriate handler based on response type.
func (s *Signal) awaitAndHandleApproval(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	l, _ := log.WorkflowLogger(ctx)

	if err := writeDirective(ctx, step.ID, DirectiveAwaitApproval, map[string]any{
		"step_idx": step.Idx,
		"status":   "awaiting-approval",
	}); err != nil {
		return errors.Wrap(err, "unable to write await-approval directive")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_idx":   step.Idx,
				"status":     "ok",
				DirectiveKey: DirectiveAwaitApproval,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to awaiting approval status")
	}

	resp, err := s.waitForApprovalResponse(ctx, flw, step)
	if err != nil {
		return err
	}

	switch resp.Type {
	case app.WorkflowStepApprovalResponseTypeApprove:
		return s.handleApproveResponse(ctx, l, step, flw)
	case app.WorkflowStepApprovalResponseTypeSkipCurrent:
		return s.handleSkipResponse(ctx, l, step, flw)
	case app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents:
		return s.handleSkipDependentsResponse(ctx, l, step, flw)
	default:
		return s.handleDenyResponse(ctx, l, step, flw)
	}
}

// waitForApprovalResponse waits for an approval response reactively using the
// "approve-plan" update handler.
func (s *Signal) waitForApprovalResponse(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) (*app.WorkflowStepApprovalResponse, error) {
	// Wait for approve-plan update handler to set s.approved (30-day deadline)
	ok, err := workflow.AwaitWithTimeout(ctx, 30*24*time.Hour, func() bool {
		return s.approved
	})
	if err != nil {
		return nil, fmt.Errorf("error waiting for approval for step %s: %w", step.ID, err)
	}
	if !ok {
		statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalExpired, map[string]any{
				"err_message": "approval was not accepted",
			}),
		})
		return nil, fmt.Errorf("approval timed out for step %s", step.ID)
	}

	// Fetch the step from DB to get the approval response.
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step after approval")
	}
	if step.Approval == nil || step.Approval.Response == nil {
		return nil, errors.New("approval response not found after update")
	}
	return step.Approval.Response, nil
}
