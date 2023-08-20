package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, appID string, dryRun bool) error {
	// update status
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		AppID:             appID,
		Status:            "deprovisioning",
		StatusDescription: "deleting application resources",
	}); err != nil {
		return fmt.Errorf("unable to update app status: %w", err)
	}

	// call child workflow
	if dryRun {
		workflow.Sleep(ctx, time.Second*5)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		AppID: appID,
	}); err != nil {
		return fmt.Errorf("unable to delete app: %w", err)
	}
	return nil
}
