package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) Created(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// send a created notification
	if install.InstallNumber == 1 {
		w.sendNotification(ctx, notifications.NotificationsTypeFirstInstallCreated, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})
	}

	if install.InstallNumber > 1 {
		w.sendNotification(ctx, notifications.NotificationsTypeInstallCreated, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})

		w.analytics.Track(ctx, events.InstallCreated, map[string]any{
			"install_id": install.ID,
		})
	}

	return w.ensureSubLoops(ctx, install, sreq)
}
