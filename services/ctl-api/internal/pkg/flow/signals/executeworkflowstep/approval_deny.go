package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// handleDenyResponse processes a deny (default) response.
func (s *Signal) handleDenyResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: denied",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if od, ok := stepSignal(step).(signal.SignalWithOnDeny); ok {
		if err := od.OnDeny(ctx); err != nil {
			l.Warn("OnDeny hook failed", zap.Error(err))
		}
	}

	// Write the stop directive so it bubbles up through step-finished → group → flow.
	if err := setResultDirective(ctx, step.ID, DirectiveStop); err != nil {
		return errors.Wrap(err, "unable to set result directive for denied step")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalDenied, map[string]any{
			"reason":     "approval denied",
			DirectiveKey: DirectiveStop,
		}),
	}); err != nil {
		return errors.Wrap(err, "unable to update")
	}
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	return nil
}
