package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

const (
	defaultTerraformVersion string = "v1.5.3"
)

func (w *Workflows) syncConfig(ctx workflow.Context, appID, appConfigID string, dryRun bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, StatusError, "org is unhealthy")
		return err
	}

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	var appCfg app.AppConfig
	if err := w.defaultExecGetActivity(ctx, w.acts.GetAppConfig, activities.GetAppConfigRequest{
		AppConfigID: appConfigID,
	}, &appCfg); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to get app config from database")
		return fmt.Errorf("unable to get app config from database: %w", err)
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusSyncing, "syncing description")
	if err := w.defaultExecErrorActivity(ctx, w.acts.SyncAppMetadata, activities.SyncAppMetadataRequest{
		AppConfigID: appConfigID,
	}); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to sync app metadata")
		return fmt.Errorf("unable to sync app metadata: %w", err)
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusSyncing, "applying terraform config")
	_, err := w.execSyncWorkflow(ctx, dryRun, &appsv1.SyncRequest{
		OrgId:            currentApp.OrgID,
		AppId:            appID,
		AppConfigId:      appConfigID,
		TerraformJson:    appCfg.GeneratedTerraform,
		TerraformVersion: defaultTerraformVersion,
		ApiToken:         currentApp.CreatedBy.Token,
		ApiUrl:           w.cfg.AppSyncAPIURL,
	})
	if err != nil {
		w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusError, "unable to sync app config")
		w.sendNotification(ctx, notifications.NotificationsTypeAppSyncError, appID, map[string]string{
			"app_name":   currentApp.Name,
			"created_by": currentApp.CreatedBy.Email,
		})
		return fmt.Errorf("unable to provision app: %w", err)
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusActive, "synced")
	w.sendNotification(ctx, notifications.NotificationsTypeAppSyncOK, appID, map[string]string{
		"app_name":   currentApp.Name,
		"created_by": currentApp.CreatedBy.Email,
	})
	return nil
}
