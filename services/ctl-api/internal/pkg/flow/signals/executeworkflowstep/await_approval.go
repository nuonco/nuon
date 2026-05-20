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

// awaitApprovalResponse writes the await-approval directive, waits for a user
// response, and returns it. The caller is responsible for dispatching the response
// to the appropriate handler.
func (s *Signal) awaitApprovalResponse(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (*app.WorkflowStepApprovalResponse, error) {
	if err := setResultDirective(ctx, step.ID, DirectiveAwaitApproval); err != nil {
		return nil, errors.Wrap(err, "unable to write await-approval directive")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_idx":           step.Idx,
				"status":             "ok",
				string(DirectiveKey): string(DirectiveAwaitApproval),
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to update step to awaiting approval status")
	}

	// Update workflow status so the UI shows the workflow is awaiting approval.
	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_id": step.ID,
			},
		},
	})

	return s.waitForApprovalResponse(ctx, flw, step)
}

// dispatchApprovalResponse routes the approval response to the appropriate handler.
func (s *Signal) dispatchApprovalResponse(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) error {
	l, _ := log.WorkflowLogger(ctx)

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
	ok, err := workflow.AwaitWithTimeout(ctx, 30*24*time.Hour, func() bool {
		return s.approved || s.retried || s.canceled || s.skipped
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

	if s.retried || s.canceled || s.skipped {
		return nil, nil
	}

	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step after approval")
	}
	if step.Approval == nil || step.Approval.Response == nil {
		return nil, errors.New("approval response not found after update")
	}
	return step.Approval.Response, nil
}

// awaitAndHandleApproval is the legacy entry point that combines awaiting and
// dispatching. Kept for backward compatibility but new code should use
// awaitApprovalResponse + dispatchApprovalResponse separately.
func (s *Signal) awaitAndHandleApproval(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	l, _ := log.WorkflowLogger(ctx)
	_ = l

	resp, err := s.awaitApprovalResponse(ctx, step, flw)
	if err != nil {
		return err
	}

	if s.retried || s.canceled {
		return nil
	}

	return s.dispatchApprovalResponse(ctx, step, flw, resp)
}
