package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) created(ctx workflow.Context, installID string) error {
	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get app: %w", err)
	}

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
	}
	return nil
}
