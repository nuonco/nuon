package worker

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTerraformVersion string = "v1.5.3"
)

func (w *Workflows) syncConfig(ctx workflow.Context, appID, appConfigID string, dryRun bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, StatusError, "org is unhealthy")
		return err
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusSyncing, "generating terraform config")

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	var generateResponse activities.GenerateTerraformConfigResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.GenerateTerraformConfig, activities.GenerateTerraformConfigRequest{
		AppConfigID: appConfigID,
	}, &generateResponse); err != nil {
		w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusError, "unable to generate terraform config - "+err.Error())
		return fmt.Errorf("unable to generate terraform config: %w", err)
	}

	if err := w.defaultExecErrorActivity(ctx, w.acts.SaveTerraformConfig, activities.SaveTerraformConfigRequest{
		AppConfigID:   appConfigID,
		TerraformJSON: generateResponse.Byts,
	}); err != nil {
		w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusError, "unable to generate terraform config - "+err.Error())
		return fmt.Errorf("unable to generate terraform config: %w", err)
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusSyncing, "applying terraform config")
	_, err := w.execSyncWorkflow(ctx, dryRun, &appsv1.SyncRequest{
		OrgId:            currentApp.OrgID,
		AppId:            appID,
		AppConfigId:      appConfigID,
		TerraformJson:    string(generateResponse.Byts),
		TerraformVersion: defaultTerraformVersion,
		ApiToken:         currentApp.CreatedBy.Token,
		ApiUrl:           w.cfg.AppSyncAPIURL,
	})
	if err != nil {
		w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusError, "unable to sync app config")
		return fmt.Errorf("unable to provision app: %w", err)
	}

	w.updateConfigStatus(ctx, appConfigID, app.AppConfigStatusActive, "synced")
	return nil
}
