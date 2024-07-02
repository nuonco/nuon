package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) provision(ctx workflow.Context, appID string, dryRun bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "org is unhealthy")
		return err
	}

	w.updateStatus(ctx, appID, app.AppStatusProvisioning, "provisioning app resources")

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &appsv1.ProvisionRequest{
		OrgId: currentApp.OrgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to provision app")
		return fmt.Errorf("unable to provision app: %w", err)
	}

	// update status with response
	w.updateStatus(ctx, appID, app.AppStatusActive, "app resources are provisioned")
	return nil
}
