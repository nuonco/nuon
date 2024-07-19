package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, appID string, status app.AppStatus, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		AppID:             appID,
		Status:            status,
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
