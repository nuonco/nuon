package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

func (w *Workflows) created(ctx workflow.Context, appID string) error {
	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeAppCreated, appID, map[string]string{
		"app_name":   currentApp.Name,
		"created_by": currentApp.CreatedBy.Email,
	})
	return nil
}
