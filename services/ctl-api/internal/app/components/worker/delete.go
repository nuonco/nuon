package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	releases "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) pollChildrenDeprovisioned(ctx workflow.Context, compID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)

	inFlight := true
	for inFlight {
		var comp app.Component
		if err := w.defaultExecGetActivity(ctx, w.acts.GetComponent, activities.GetComponentRequest{
			ComponentID: compID,
		}, &comp); err != nil {
			w.updateStatus(ctx, compID, "error", "unable to get component from database")
			return fmt.Errorf("unable to get component: %w", err)
		}

		inFlight = false
		for _, cfgVersion := range comp.ComponentConfigs {
			for _, bld := range cfgVersion.ComponentBuilds {
				for _, rel := range bld.ComponentReleases {
					if rel.Status == string(releases.StatusActive) || rel.Status == string(releases.StatusError) || rel.Status == string(releases.StatusFailed) {
						continue
					}

					inFlight = true
				}
			}
		}

		if workflow.Now(ctx).After(deadline) {
			w.updateStatus(ctx, compID, "error", "time out waiting for releases to finish")
			return fmt.Errorf("timeout waiting for installs to deprovision")
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

func (w *Workflows) delete(ctx workflow.Context, componentID string, dryRun bool) error {
	w.updateStatus(ctx, componentID, StatusActive, "polling for releases to finish")
	if err := w.pollChildrenDeprovisioned(ctx, componentID); err != nil {
		return err
	}

	// update status
	w.updateStatus(ctx, componentID, StatusDeprovisioning, "deleting component")
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		ComponentID: componentID,
	}); err != nil {
		return fmt.Errorf("unable to delete component: %w", err)
	}

	return nil
}
