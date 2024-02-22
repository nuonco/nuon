package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) forceDelete(ctx workflow.Context, orgID string, dryRun bool) error {
	l := workflow.GetLogger(ctx)
	if err := w.deprovision(ctx, orgID, dryRun); err != nil {
		l.Error("unable to deprovision org: %w", zap.Error(err))
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		OrgID: orgID,
	}); err != nil {
		l.Error("unable to delete org: %w", zap.Error(err))
	}
	return nil
}
