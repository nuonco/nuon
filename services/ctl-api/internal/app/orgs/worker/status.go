package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
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

func (w *Workflows) updateStatus(ctx workflow.Context, orgID string, status Status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            string(status),
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	w.l.Error("unable to update org status",
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

	w.l.Error("unable to update org health check status",
		zap.String("health-check-id", orgHealthCheckID),
		zap.Error(err))
}
