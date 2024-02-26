package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusProvisioning   Status = "provisioning"
	StatusDeprovisioning Status = "deprovisioning"
	StatusActive         Status = "active"
	StatusError          Status = "error"
)

func (w *Workflows) updateStatus(ctx workflow.Context, appID string, status Status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		AppID:             appID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update status",
		zap.String("app-id", appID),
		zap.Error(err))
}

func (w *Workflows) updateConfigStatus(ctx workflow.Context, appConfigID string, status app.AppConfigStatus, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateConfigStatus, activities.UpdateConfigStatusRequest{
		AppConfigID:       appConfigID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update app config status",
		zap.String("app-config-id", appConfigID),
		zap.Error(err))
}
