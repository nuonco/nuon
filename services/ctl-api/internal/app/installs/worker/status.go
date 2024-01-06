package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusPlanning       Status = "planning"
	StatusError          Status = "error"
	StatusAccessError    Status = "access-error"
	StatusActive         Status = "active"
	StatusProvisioning   Status = "provisioning"
	StatusDeprovisioning Status = "deprovisioning"
	StatusDeprovisioned  Status = "deprovisioned"

	StatusSyncing   Status = "syncing"
	StatusExecuting Status = "executing"
)

func (w *Workflows) updateStatus(ctx workflow.Context, installID string, status Status, statusDescription string) {
	l := workflow.GetLogger(ctx)

	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		InstallID:         installID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update install status",
		zap.String("install-id", installID),
		zap.Error(err))
}

func (w *Workflows) updateRunStatus(ctx workflow.Context, runID string, status Status, statusDescription string) {
	l := workflow.GetLogger(ctx)

	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateRunStatus, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update run status",
		zap.String("run-id", runID),
		zap.Error(err))
}

func (w *Workflows) updateDeployStatus(ctx workflow.Context, deployID string, status Status, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update deploy status",
		zap.String("deploy-id", deployID),
		zap.Error(err))
}
