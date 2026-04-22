package executeworkflowstep

import (
	"strconv"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// autoApprovalCheck builds the auto-approval plan check.
// shouldRun is always true — processPlan only runs for approval steps, so we
// always evaluate auto-approval. The run function checks the workflow's
// ApprovalOption and short-circuits if approve-all is set.
func (s *Signal) autoApprovalCheck(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) planCheck {
	return planCheck{
		name: "auto-approval",
		shouldRun: func() bool {
			return true
		},
		run: func() (bool, error) {
			return s.runAutoApprovalCheck(ctx, step, flw)
		},
	}
}

// runAutoApprovalCheck auto-approves the step when the install has approve-all set.
// Returns done=false (no short-circuit) when the option is not approve-all.
func (s *Signal) runAutoApprovalCheck(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (bool, error) {
	if flw.ApprovalOption != app.InstallApprovalOptionApproveAll {
		return false, nil
	}

	// Create approval response
	_, err := activities.AwaitCreateApprovalResponse(ctx, activities.CreateStepApprovalResponseRequest{
		StepApprovalID: step.Approval.ID,
		Type:           app.WorkflowStepApprovalResponseTypeApprove,
		Note:           "auto-approved",
	})
	if err != nil {
		return false, errors.Wrap(err, "unable to create auto-approval response")
	}

	// Update step status
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto-approved " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx":      step.Idx,
				"status":        "auto-approved",
				"auto_approved": true,
			},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step status for auto-approval")
	}

	// Update flow status
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "auto-approved step " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "auto-approved",
			},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update flow status for auto-approval")
	}

	// Call OnApprove hook if the signal implements it
	if oa, ok := stepSignal(step).(signal.SignalWithOnApprove); ok {
		if err := oa.OnApprove(ctx); err != nil {
			l, _ := log.WorkflowLogger(ctx)
			if l != nil {
				l.Warn("OnApprove hook failed during auto-approval", zap.Error(err))
			}
		}
	}

	// Write continue directive so the group proceeds to the next step (e.g. apply).
	if err := setResultDirective(ctx, step.ID, DirectiveContinue); err != nil {
		return false, errors.Wrap(err, "unable to set result directive for auto-approval")
	}

	return true, nil
}
