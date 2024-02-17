package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusError          Status = "error"
	StatusActive         Status = "active"
	StatusDeprovisioning Status = "deprovisioning"
)

func (w *Workflows) updateStatus(ctx workflow.Context, compID string, status Status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		ComponentID:       compID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update org status",
		zap.String("organization-id", compID),
		zap.Error(err))
}
