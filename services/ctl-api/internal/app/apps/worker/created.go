package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) Created(ctx workflow.Context, sreq signals.RequestSignal) error {
	appID := sreq.ID

	currentApp, err := activities.AwaitGetByAppID(ctx, appID)
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeAppCreated, appID, map[string]string{
		"app_name":   currentApp.Name,
		"created_by": currentApp.CreatedBy.Email,
	})

	w.analytics.Track(ctx, analytics.AppCreated, map[string]interface{}{
		"app_id": appID,
	})

	return nil
}
