package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) pollInstallsDeprovisioned(ctx workflow.Context, appID string) error {
	for {
		var currentApp app.App
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			AppID: appID,
		}, &currentApp); err != nil {
			w.updateStatus(ctx, appID, "error", "unable to get app from database")
			return fmt.Errorf("unable to get app: %w", err)
		}

		if len(currentApp.Installs) < 1 {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

func (w *Workflows) deprovision(ctx workflow.Context, appID string, dryRun bool) error {
	w.updateStatus(ctx, appID, StatusDeprovisioning, "polling for all installs to be deprovisioned")
	if err := w.pollInstallsDeprovisioned(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, StatusError, "error polling installs being deprovisioned")
		return fmt.Errorf("unable to poll for deleted installs: %w", err)
	}

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

	return nil
}
