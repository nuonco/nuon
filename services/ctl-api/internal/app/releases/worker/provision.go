package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
)

func (w *Workflows) provision(ctx workflow.Context, releaseID string, dryRun bool) error {
	// TODO(ja): "provisioning" as a status for releases doesn't sound right.
	// We may need to revisit release statuses.
	w.updateStatus(ctx, releaseID, app.ReleaseStatusProvisioning, "provisioning deploys")

	var release app.ComponentRelease
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		ReleaseID: releaseID,
	}, &release); err != nil {
		w.updateStatus(ctx, releaseID, "error", "unable to read release record from database")
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
			w.updateStatus(ctx, releaseID, "error", "release step errored")
			return fmt.Errorf("release failed: %w", err)
		}
	}

	// update release status
	w.updateStatus(ctx, releaseID, "active", "release succeeded")
	return nil
}
