package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) updateRunStatus(ctx workflow.Context, runID string, status app.SandboxRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateRunStatus, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update run status",
		zap.String("run-id", runID),
		zap.Error(err))
}

func (w *Workflows) updateDeployStatus(ctx workflow.Context, deployID string, status app.InstallDeployStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	l.Error("unable to update deploy status",
		zap.String("deploy-id", deployID),
		zap.Error(err))
}
