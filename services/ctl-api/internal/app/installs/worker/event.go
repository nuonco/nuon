package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) writeInstallEvent(ctx workflow.Context, installID string, op signals.Operation, status app.OperationStatus) {
	l := workflow.GetLogger(ctx)
	if err := w.defaultExecErrorActivity(ctx, w.acts.WriteEvent, &activities.WriteEventRequest{
		InstallID:       installID,
		Operation:       op,
		OperationStatus: status,
	}); err != nil {
		l.Error("unable to write install event", zap.Error(err))
	}
}

func (w *Workflows) writeDeployEvent(ctx workflow.Context, deployID string, op signals.Operation, status app.OperationStatus) {
	l := workflow.GetLogger(ctx)
	if err := w.defaultExecErrorActivity(ctx, w.acts.WriteEvent, &activities.WriteEventRequest{
		DeployID:        deployID,
		Operation:       op,
		OperationStatus: status,
	}); err != nil {
		l.Error("unable to write deploy event", zap.Error(err))
	}
}

func (w *Workflows) writeRunEvent(ctx workflow.Context, runID string, op signals.Operation, status app.OperationStatus) {
	l := workflow.GetLogger(ctx)
	if err := w.defaultExecErrorActivity(ctx, w.acts.WriteEvent, &activities.WriteEventRequest{
		SandboxRunID:    runID,
		Operation:       op,
		OperationStatus: status,
	}); err != nil {
		l.Error("unable to write run event", zap.Error(err))
	}
}
