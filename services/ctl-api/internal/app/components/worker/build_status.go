package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Status string

const (
	StatusPlanning Status = "planning"
	StatusError    Status = "error"
	StatusBuilding Status = "building"
	StatusActive   Status = "active"
	StatusDeleting Status = "deleting"
)

func (w *Workflows) updateBuildStatus(ctx workflow.Context, bldID string, status Status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateBuildStatus, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	w.l.Error("unable to update build status",
		zap.String("build-id", bldID),
		zap.Error(err))
}
