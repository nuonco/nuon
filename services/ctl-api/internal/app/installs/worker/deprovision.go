package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

func (w *Workflows) isDeprovisionable(ctx workflow.Context, install *app.Install) (bool, error) {
	if len(install.InstallSandboxRuns) < 1 {
		return false, nil
	}

	if install.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError {
		return false, nil
	}

	untornCmpIds, err := activities.AwaitFetchUntornInstallDeploys(ctx, activities.FetchUntornInstallDeploysRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return false, fmt.Errorf("unable to fetch untorn install deploys: %w", err)
	}

	if len(untornCmpIds) > 0 {
		return false, nil
	}

	return true, nil
}

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	sandboxMode := sreq.SandboxMode

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeDeprovision,
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}

	isDeprovisionable, err := w.isDeprovisionable(ctx, install)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to determine if install is deprovisionable")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to determine if install is deprovisionable: %w", err)
	}

	if !isDeprovisionable {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "install is not deprovisionable")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusNoop)
		return nil
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusDeprovisioning, "deprovisioning")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusStarted)

	// deprovision the runner
	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDeprovision,
	})

	// wait for the runner
	if err := w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeDestroy, sandboxMode); err != nil {
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
		return err
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "successfully deprovisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFinished)
	return nil
}
