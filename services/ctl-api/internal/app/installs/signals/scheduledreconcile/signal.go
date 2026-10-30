package scheduledreconcile

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "scheduled-reconcile"

// Signal fires a user-scheduled reconcile-app-version for an install. The
// emitter that creates this signal is named with the
// installhelpers.ScheduledReconcileEmitterPrefix prefix so config-pin
// reconciliation (appconfigupdated.Signal) never sweeps it up.
type Signal struct {
	InstallID            string `json:"install_id"`
	ScheduledReconcileID string `json:"scheduled_reconcile_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "scheduled-reconcile",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.ScheduledReconcileID == "" {
		return fmt.Errorf("scheduled_reconcile_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	schedule, err := activities.AwaitGetInstallScheduledReconcile(ctx, activities.GetInstallScheduledReconcileRequest{
		ScheduledReconcileID: s.ScheduledReconcileID,
	})
	if err != nil {
		return fmt.Errorf("unable to get scheduled reconcile: %w", err)
	}

	if schedule.Status != app.InstallScheduledReconcileStatusPending {
		l.Info("scheduled reconcile is no longer pending, skipping",
			"scheduled_reconcile_id", s.ScheduledReconcileID,
			"status", string(schedule.Status))
		return nil
	}

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		// The install no longer exists (deleted/forgotten): cancel the schedule
		// rather than failing/retrying this signal forever.
		return s.markStatus(ctx, l, app.InstallScheduledReconcileStatusCancelled, "install no longer exists", nil)
	}

	if install.LifecyclePhase.Phase == lifecyclephase.Deprovisioned || install.LifecyclePhase.Phase == lifecyclephase.Deprovisioning {
		return s.markStatus(ctx, l, app.InstallScheduledReconcileStatusCancelled, "install is deprovisioned", nil)
	}

	disabled, err := activities.AwaitIsRunnerDisabled(ctx, activities.IsRunnerDisabledRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to check whether the install runner is disabled: %w", err)
	}
	if disabled.Disabled {
		return s.markStatus(ctx, l, app.InstallScheduledReconcileStatusFailed, "install runner is disabled", nil)
	}

	result, err := activities.AwaitFireInstallScheduledReconcile(ctx, activities.FireInstallScheduledReconcileRequest{
		InstallID:            s.InstallID,
		ScheduledReconcileID: s.ScheduledReconcileID,
		PlanOnly:             schedule.PlanOnly,
	})
	if err != nil {
		return fmt.Errorf("unable to fire scheduled reconcile: %w", err)
	}

	var workflowID *string
	if result.WorkflowID != "" {
		workflowID = &result.WorkflowID
	}
	return s.markStatus(ctx, l, result.Status, result.Reason, workflowID)
}

func (s *Signal) markStatus(ctx workflow.Context, l log.Logger, status app.InstallScheduledReconcileStatus, reason string, workflowID *string) error {
	if err := activities.AwaitMarkInstallScheduledReconcileStatus(ctx, activities.MarkInstallScheduledReconcileStatusRequest{
		ScheduledReconcileID: s.ScheduledReconcileID,
		Status:               status,
		StatusReason:         reason,
		WorkflowID:           workflowID,
	}); err != nil {
		return fmt.Errorf("unable to update scheduled reconcile status: %w", err)
	}
	return nil
}
