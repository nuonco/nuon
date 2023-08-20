package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deprovision(ctx workflow.Context, orgID string, dryRun bool) error {
	// update status
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		OrgID:             orgID,
		Status:            "deprovisioning",
		StatusDescription: "tearing down org server, runner and infra - this should take about 15 minutes",
	}); err != nil {
		return fmt.Errorf("unable to update org status: %w", err)
	}

	// call child workflow
	if dryRun {
		workflow.Sleep(ctx, time.Second*5)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		OrgID: orgID,
	}); err != nil {
		return fmt.Errorf("unable to delete org: %w", err)
	}
	return nil
}
