package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) updateStatus(ctx workflow.Context, releaseID, status, statusDescription string) {
	err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		ReleaseID:         releaseID,
		Status:            status,
		StatusDescription: statusDescription,
	})
	if err == nil {
		return
	}

	w.l.Error("unable to update release status",
		zap.String("release-id", releaseID),
		zap.Error(err))
}

func (w *Workflows) provision(ctx workflow.Context, releaseID string, dryRun bool) error {
	w.updateStatus(ctx, releaseID, "provisioning", "provisioning release")

	var release app.ComponentRelease
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		ReleaseID: releaseID,
	}, &release); err != nil {
		w.updateStatus(ctx, releaseID, "failed", "unable to read release record from database")
		return fmt.Errorf("unable to read release record from database: %w", err)
	}

	w.updateStatus(ctx, releaseID, "executing_release_steps", fmt.Sprintf("executing %d release steps", len(release.ComponentReleaseSteps)))

	// now trigger each step of the release
	for _, step := range release.ComponentReleaseSteps {
		stepWorkflowID := provisionStepWorkflowID(releaseID, step.ID)
		req := ProvisionReleaseStepRequest{
			ReleaseID:     releaseID,
			ReleaseStepID: step.ID,
		}
		if _, err := w.execProvisionStepWorkflow(ctx,
			stepWorkflowID,
			req,
		); err != nil {
			w.updateStatus(ctx, releaseID, "failed", "release step failed")
			return fmt.Errorf("release failed: %w", err)
		}
	}

	// update release status
	w.updateStatus(ctx, releaseID, "active", "release succeeded")
	return nil
}
