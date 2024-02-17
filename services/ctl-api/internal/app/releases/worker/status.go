package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusPlanning       Status = "planning"
	StatusError          Status = "error"
	StatusActive         Status = "active"
	StatusProvisioning   Status = "provisioning"
	StatusDeprovisioning Status = "deprovisioning"

	StatusSyncing   Status = "syncing"
	StatusExecuting Status = "executing"
)

func (w *Workflows) updateStatus(ctx workflow.Context, releaseID string, status Status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		ReleaseID:         releaseID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update release status",
		zap.String("release-id", releaseID),
		zap.Error(err))
}
