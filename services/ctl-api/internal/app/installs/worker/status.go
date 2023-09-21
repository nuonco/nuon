package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusPlanning	     Status = "planning"
	StatusError	     Status = "error"
	StatusActive	     Status = "active"
	StatusProvisioning   Status = "provisioning"
	StatusDeprovisioning Status = "deprovisioning"

	StatusSyncing	Status = "syncing"
	StatusExecuting Status = "executing"
)

func (w *Workflows) updateStatus(ctx workflow.Context, installID string, status Status, statusDescription string) {
	l := workflow.GetLogger(ctx)

	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		InstallID:	   installID,
		Status:		   string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update install status",
		zap.String("install-id", installID),
		zap.Error(err))
}

func (w *Workflows) updateDeployStatus(ctx workflow.Context, deployID string, status Status, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:	   deployID,
		Status:		   string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update deploy status",
		zap.String("deploy-id", deployID),
		zap.Error(err))
}
