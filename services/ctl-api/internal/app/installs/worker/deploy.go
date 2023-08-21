package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deploy(ctx workflow.Context, installID, deployID string, dryRun bool) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            "planning",
		StatusDescription: "creating deploy plan",
	}); err != nil {
		return fmt.Errorf("unable to update deploy status: %w", err)
	}

	// execute the plan phase here
	if dryRun {
		workflow.Sleep(ctx, w.cfg.DevDryRunSleep)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            "deploying",
		StatusDescription: "executing deploy plan",
	}); err != nil {
		return fmt.Errorf("unable to update deploy status: %w", err)
	}

	// execute the exec phase here
	if dryRun {
		workflow.Sleep(ctx, w.cfg.DevDryRunSleep)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateDeployStatus, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            "active",
		StatusDescription: "active",
	}); err != nil {
		return fmt.Errorf("unable to update deploy status: %w", err)
	}

	return nil
}
