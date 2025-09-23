package worker

import (
	"fmt"
	"slices"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
)

func (w *Workflows) pollComponentBeingUnused(ctx workflow.Context, compID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)

	inFlight := true
	for inFlight {
		appCfg, err := activities.AwaitGetComponentAppConfigByComponentID(ctx, compID)
		if err != nil {
			w.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component app config")
			return fmt.Errorf("unable to get component app config: %w", err)
		}

		inFlight = false
		if slices.Contains(appCfg.ComponentIDs, compID) {
			inFlight = true
		}

		depsInstalls, err := activities.AwaitGetComponentInstalls(ctx, activities.GetComponentInstallsRequest{
			ComponentID: compID,
			AppID:       appCfg.AppID,
		})
		if err != nil {
			w.updateStatus(ctx, compID, app.ComponentStatusError, "delete: unable to get component dependent installs")
			return fmt.Errorf("unable to get component dependent installs: %w", err)
		}

		if len(depsInstalls) > 0 {
			inFlight = true
		}

		if workflow.Now(ctx).After(deadline) {
			w.updateStatus(ctx, compID, app.ComponentStatusError, "delete: timed out waiting for dependent installs to remove the component")
			return fmt.Errorf("timeout waiting for dependent installs to remove the component")
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.ComponentStatusDeleteQueued, "delete: polling for component being unused by active installs and latest app config")
	if err := w.pollComponentBeingUnused(ctx, sreq.ID); err != nil {
		return err
	}

	// update status
	w.updateStatus(ctx, sreq.ID, app.ComponentStatusDeprovisioning, "deleting component")
	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		ComponentID: sreq.ID,
	}); err != nil {
		w.updateStatus(ctx, sreq.ID, app.ComponentStatusError, "unable to delete component from database")
		return fmt.Errorf("unable to delete component: %w", err)
	}

	return nil
}
