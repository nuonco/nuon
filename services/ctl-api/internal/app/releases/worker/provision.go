package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	// TODO(ja): "provisioning" as a status for releases doesn't sound right.
	// We may need to revisit release statuses.
	w.updateStatus(ctx, sreq.ID, app.ReleaseStatusProvisioning, "provisioning deploys")

	release, err := activities.AwaitGetByReleaseID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.ReleaseStatusError, "unable to read release record from database")
		return fmt.Errorf("unable to read release record from database: %w", err)
	}

	w.updateStatus(ctx, sreq.ID, "executing_release_steps", fmt.Sprintf("executing %d release steps", len(release.ComponentReleaseSteps)))

	// now trigger each step of the release
	for _, step := range release.ComponentReleaseSteps {
		stepWorkflowID := provisionStepWorkflowID(sreq.ID, step.ID)
		req := ProvisionReleaseStepRequest{
			ReleaseID:     sreq.ID,
			ReleaseStepID: step.ID,
		}
		if _, err := w.execProvisionStepWorkflow(ctx,
			stepWorkflowID,
			req,
		); err != nil {
			w.updateStatus(ctx, sreq.ID, "error", "release step errored")
			return fmt.Errorf("release failed: %w", err)
		}
	}

	// update release status
	w.updateStatus(ctx, sreq.ID, "active", "release succeeded")

	app, err := activities.AwaitGetReleaseAppByReleaseID(ctx, sreq.ID)
	w.sendNotification(ctx, notifications.NotificationsTypeReleaseSucceeded, sreq.ID, map[string]string{
		"app_name":   app.Name,
		"created_by": release.CreatedBy.Email,
	})
	return nil
}
