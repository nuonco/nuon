package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

func (w *Workflows) updateStatus(ctx workflow.Context, orgID string, status app.OrgStatus, statusDescription string) {
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l := workflow.GetLogger(ctx)
		l.Error("unable to update org status",
			zap.String("organization-id", orgID),
			zap.Error(err))
	}
}
