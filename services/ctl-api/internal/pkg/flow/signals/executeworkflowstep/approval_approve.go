package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// handleApproveResponse processes an "approve" response.
func (s *Signal) handleApproveResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: approved",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if oa, ok := stepSignal(step).(signal.SignalWithOnApprove); ok {
		if err := oa.OnApprove(ctx); err != nil {
			l.Warn("OnApprove hook failed", zap.Error(err))
		}
	}

	// Workflow resumes from approval-awaiting.
	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "approved " + step.Name + ", continuing",
		},
	})

	return writeDirective(ctx, step.ID, DirectiveContinue, map[string]any{
		"step_idx": step.Idx,
		"status":   "approved",
	})
}
