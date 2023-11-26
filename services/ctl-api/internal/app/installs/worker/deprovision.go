package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, installID string, dryRun bool) error {
	w.updateStatus(ctx, installID, StatusDeprovisioning, "deprovisioning install resources")

	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	req, err := w.protos.ToInstallDeprovisionRequest(&install)
	if err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to create install deprovision request")
		return fmt.Errorf("unable to create install deprovision request: %w", err)
	}

	_, err = w.execDeprovisionWorkflow(ctx, dryRun, req)
	if err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to deprovision install resources")
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	return nil

}

func (w *Workflows) delete(ctx workflow.Context, installID string, dryRun bool) error {
	if err := w.deprovision(ctx, installID, dryRun); err != nil {
		return err
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
