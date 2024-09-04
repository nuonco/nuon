package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
)

func (w *Workflows) build(ctx workflow.Context, cmpID, buildID string, sandboxMode bool) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "creating build plan")

	currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, cmpID)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, cmpID)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	if comp.Status != app.ComponentStatusActive {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "component is not active")
		return fmt.Errorf("component is not active")
	}

	if currentApp.Org.OrgType != app.OrgTypeV2 {
		if err := w.execBuildLegacy(ctx, cmpID, buildID, currentApp, sandboxMode); err != nil {
			return err
		}

		return nil
	}

	if err := w.execBuild(ctx, cmpID, buildID, currentApp, sandboxMode); err != nil {
		return err
	}

	return nil
}
