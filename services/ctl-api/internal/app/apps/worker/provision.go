package worker

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, appID string, dryRun bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, StatusError, "org is unhealthy")
		return err
	}

	w.updateStatus(ctx, appID, StatusProvisioning, "provisioning app resources")

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &appsv1.ProvisionRequest{
		OrgId: currentApp.OrgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to provision app")
		return fmt.Errorf("unable to provision app: %w", err)
	}

	// update status with response
	w.updateStatus(ctx, appID, StatusActive, "app resources are provisioned")
	return nil
}
