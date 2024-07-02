package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, orgID string, status app.OrgStatus, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update org status",
		zap.String("organization-id", orgID),
		zap.Error(err))
}

func (w *Workflows) updateHealthCheckStatus(ctx workflow.Context, orgHealthCheckID string, status app.OrgHealthCheckStatus, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateHealthCheckStatus, activities.UpdateHealthCheckStatusRequest{
		OrgHealthCheckID:  orgHealthCheckID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l := workflow.GetLogger(ctx)
	l.Error("unable to update org health check status",
		zap.String("health-check-id", orgHealthCheckID),
		zap.Error(err))
}
