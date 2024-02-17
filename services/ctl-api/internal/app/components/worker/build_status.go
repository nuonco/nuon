package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type BuildStatus string

const (
	BuildStatusPlanning BuildStatus = "planning"
	BuildStatusError    BuildStatus = "error"
	BuildStatusBuilding BuildStatus = "building"
	BuildStatusActive   BuildStatus = "active"
	BuildStatusDeleting BuildStatus = "deleting"
)

func (w *Workflows) updateBuildStatus(ctx workflow.Context, bldID string, status BuildStatus, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateBuildStatus, activities.UpdateBuildStatus{
		BuildID:           bldID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update build status",
		zap.String("build-id", bldID),
		zap.Error(err))
}
