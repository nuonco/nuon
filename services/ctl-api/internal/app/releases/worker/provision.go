package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, releaseID string, dryRun bool) error {
	var release app.ComponentRelease
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		ReleaseID: releaseID,
	}, &release); err != nil {
		return fmt.Errorf("unable to get release: %w", err)
	}

	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		ReleaseID:         releaseID,
		Status:            "executing_release_steps",
		StatusDescription: fmt.Sprintf("executing %d release steps", len(release.ComponentReleaseSteps)),
	}); err != nil {
		return fmt.Errorf("unable to update release status: %w", err)
	}

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
			if updateErr := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
				ReleaseID:         releaseID,
				Status:            "failed",
				StatusDescription: "release step failed",
			}); updateErr != nil {
				return fmt.Errorf("unable to update release status: %w", err)
			}

			return fmt.Errorf("release failed: %w", err)
		}
	}

	// update release status
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		ReleaseID:         releaseID,
		Status:            "active",
		StatusDescription: "release succeeded",
	}); err != nil {
		return fmt.Errorf("unable to update release status: %w", err)
	}
	return nil
}
