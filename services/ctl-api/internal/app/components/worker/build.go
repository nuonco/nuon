package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Build(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusPlanning, "creating build plan")

	currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, sreq.ID)
	if err != nil {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	notify := func(err error) error {
		w.sendNotification(ctx, notifications.NotificationsTypeComponentBuildFailed, currentApp.ID, map[string]string{
			"component_name": comp.Name,
			"app_name":       currentApp.Name,
			"created_by":     currentApp.CreatedBy.Email,
		})
		return err
	}

	if comp.Status != app.ComponentStatusActive {
		w.updateBuildStatus(ctx, sreq.BuildID, app.ComponentBuildStatusError, "component is not active")
		return notify(fmt.Errorf("component is not active"))
	}

	if currentApp.Org.OrgType != app.OrgTypeV2 {
		if err := w.execBuildLegacy(ctx, sreq.ID, sreq.BuildID, currentApp, sreq.SandboxMode); err != nil {
			return notify(err)
		}

		return nil
	}

	if err := w.execBuild(ctx, sreq.ID, sreq.BuildID, currentApp, sreq.SandboxMode); err != nil {
		return notify(err)
	}

	return nil
}
