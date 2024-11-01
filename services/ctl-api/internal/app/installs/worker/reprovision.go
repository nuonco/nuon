package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Reprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	sandboxMode := sreq.SandboxMode

	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeReprovision,
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create sandbox run")
		return fmt.Errorf("unable to create install: %w", err)
	}

	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusStarted)
	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusProvisioning, "provisioning")

	tempLogStreamID, err := w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeCreate, sandboxMode)
	if err != nil {
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return err
	}

	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type:        runnersignals.OperationReprovision,
		LogStreamID: tempLogStreamID,
	})
	if err := w.pollRunner(ctx, install.RunnerGroup.Runners[0].ID); err != nil {
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return err
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "install resources provisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFinished)

	return nil
}
