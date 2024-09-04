package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

func (w *Workflows) deprovisionLegacy(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, sandboxMode bool) error {
	req, err := w.protos.ToInstallDeprovisionRequest(install, installRun.ID)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install deprovision request")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to create install deprovision request: %w", err)
	}
	_, err = w.execDeprovisionWorkflow(ctx, sandboxMode, req)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to deprovision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	return nil
}

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

func (w *Workflows) deprovision(ctx workflow.Context, installID string, sandboxMode bool) error {
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

	if install.Org.OrgType != app.OrgTypeV2 {
		if err := w.deprovisionLegacy(ctx, install, installRun, sandboxMode); err != nil {
			return err
		}

		return nil
	}

	// deprovision the runner
	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationDeprovision,
	})
	// wait for the runner

	if err := w.executeSandboxRun(ctx, install, installRun, app.RunnerJobOperationTypeDestroy, sandboxMode); err != nil {
		return err
	}

	w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusActive, "successfully deprovisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusFinished)
	return nil
}

func (w *Workflows) delete(ctx workflow.Context, installID string, dryRun bool) error {
	if err := w.deprovision(ctx, installID, dryRun); err != nil {
		return err
	}

	// fail all queued deploys
	if err := activities.AwaitFailQueuedDeploysByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to fail queued install: %w", err)
	}

	// update status with response
	if err := activities.AwaitDeleteByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
