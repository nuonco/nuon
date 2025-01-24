package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) updateRunStatus(ctx workflow.Context, runID string, status app.InstallActionWorkflowRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateInstallWorkflowRunStatus(ctx, activities.UpdateInstallWorkflowRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status",
			zap.String("run-id", runID),
			zap.Error(err))
	}
}
