package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) build(ctx workflow.Context, cmpID, buildID string, dryRun bool) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateBuildStatus, activities.UpdateBuildStatus{
		BuildID:           buildID,
		Status:            "planning",
		StatusDescription: "creating build plan",
	}); err != nil {
		return fmt.Errorf("unable to update build status: %w", err)
	}

	// execute the plan phase here
	if dryRun {
		workflow.Sleep(ctx, w.cfg.DevDryRunSleep)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateBuildStatus, activities.UpdateBuildStatus{
		BuildID:           buildID,
		Status:            "building",
		StatusDescription: "executing build plan",
	}); err != nil {
		return fmt.Errorf("unable to update build status: %w", err)
	}

	// execute the exec phase here
	if dryRun {
		workflow.Sleep(ctx, w.cfg.DevDryRunSleep)
	}

	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateBuildStatus, activities.UpdateBuildStatus{
		BuildID:           buildID,
		Status:            "active",
		StatusDescription: "build is active and ready to be deployed",
	}); err != nil {
		return fmt.Errorf("unable to update build status: %w", err)
	}
	return nil
}
