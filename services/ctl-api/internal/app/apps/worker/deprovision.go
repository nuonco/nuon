package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, appID string, dryRun bool) error {
	// update status
	w.updateStatus(ctx, appID, StatusDeprovisioning, "deleting app resources")

	// NOTE: we don't actually have a deprovision step, but we sleep here locally when a dry-run is enabled, to
	// ensure that the status updating is working correctly.
	if dryRun {
		workflow.Sleep(ctx, w.cfg.DevDryRunSleep)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		AppID: appID,
	}); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to delete app")
		return fmt.Errorf("unable to delete app: %w", err)
	}

	w.updateStatus(ctx, appID, StatusActive, "app is provisioned")
	return nil
}
