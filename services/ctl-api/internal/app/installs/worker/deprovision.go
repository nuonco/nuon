package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) isDeprovisionable(install app.Install) bool {
	if len(install.InstallSandboxRuns) < 1 {
		return false
	}
	if install.InstallSandboxRuns[0].Status == string(StatusAccessError) {
		return false
	}

	return true
}

func (w *Workflows) deprovision(ctx workflow.Context, installID string, dryRun bool) error {
	w.updateStatus(ctx, installID, StatusDeprovisioning, "deprovisioning install resources")

	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installRun app.InstallSandboxRun
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateSandboxRun, activities.CreateSandboxRunRequest{
		InstallID: installID,
		RunType:   app.SandboxRunTypeDeprovision,
	}, &installRun); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to create sandbox run")
		return fmt.Errorf("unable to create install: %w", err)
	}

	if !w.isDeprovisionable(install) {
		w.updateRunStatus(ctx, installRun.ID, StatusError, "install is not deprovisionable")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationDeprovision, app.OperationStatusNoop)
		return nil
	}

	w.updateRunStatus(ctx, installRun.ID, StatusDeprovisioning, "deprovisioning")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusStarted)

	req, err := w.protos.ToInstallDeprovisionRequest(&install, installRun.ID)
	if err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to create install deprovision request")
		w.updateRunStatus(ctx, installRun.ID, StatusError, "unable to create install deprovision request")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to create install deprovision request: %w", err)
	}

	_, err = w.execDeprovisionWorkflow(ctx, dryRun, req)
	if err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to deprovision install resources")
		w.updateRunStatus(ctx, installRun.ID, StatusError, "unable to deprovision install resources")
		w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFailed)
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	w.updateStatus(ctx, installID, StatusDeprovisioned, "successfully deprovisioned")
	w.updateRunStatus(ctx, installRun.ID, StatusActive, "successfully deprovisioned")
	w.writeRunEvent(ctx, installRun.ID, signals.OperationReprovision, app.OperationStatusFinished)
	return nil
}

func (w *Workflows) delete(ctx workflow.Context, installID string, dryRun bool) error {
	if err := w.deprovision(ctx, installID, dryRun); err != nil {
		return err
	}

	// fail all queued deploys
	if err := w.defaultExecErrorActivity(ctx, w.acts.FailQueuedDeploys, activities.FailQueuedDeploysRequest{
		InstallID: installID,
	}); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to fail queued deploys database")
		return fmt.Errorf("unable to fail queued install: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		InstallID: installID,
	}); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to delete install from database")
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
