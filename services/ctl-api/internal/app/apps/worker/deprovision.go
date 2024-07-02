package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) pollChildrenDeprovisioned(ctx workflow.Context, appID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)
	for {
		var currentApp app.App
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			AppID: appID,
		}, &currentApp); err != nil {
			w.updateStatus(ctx, appID, "error", "unable to get app from database")
			return fmt.Errorf("unable to get app: %w", err)
		}

		installCnt := 0
		for _, install := range currentApp.Installs {
			// if an install was never attempted, it does not need to be polled
			if len(install.InstallSandboxRuns) < 1 {
				continue
			}

			if install.InstallSandboxRuns[0].Status != app.SandboxRunStatusAccessError {
				installCnt += 1
			}
		}

		if len(currentApp.Components) < 1 && installCnt < 1 {
			return nil
		}

		if workflow.Now(ctx).After(deadline) {
			err := fmt.Errorf("timeout waiting for installs and components to deprovision")
			w.updateStatus(ctx, appID, "error", err.Error())
			return err
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

func (w *Workflows) deprovision(ctx workflow.Context, appID string, dryRun bool) error {
	w.updateStatus(ctx, appID, app.AppStatusActive, "polling for installs and components to be deprovisioned")
	if err := w.pollChildrenDeprovisioned(ctx, appID); err != nil {
		return err
	}

	// update status
	w.updateStatus(ctx, appID, app.AppStatusDeprovisioning, "deleting app resources")

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	_, err := w.execDeprovisionWorkflow(ctx, dryRun, &appsv1.DeprovisionRequest{
		OrgId: currentApp.OrgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to deprovision app")
		return fmt.Errorf("unable to deprovision app: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		AppID: appID,
	}); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to delete app")
		return fmt.Errorf("unable to delete app: %w", err)
	}

	return nil
}
