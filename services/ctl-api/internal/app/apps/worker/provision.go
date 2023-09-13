package worker

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, appID string, dryRun bool) error {
	w.updateStatus(ctx, appID, StatusProvisioning, "creating app resources")

	if err := w.pollDependencies(ctx, appID); err != nil {
		return fmt.Errorf("unable to poll org for app: %w", err)
	}

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
	w.updateStatus(ctx, appID, StatusActive, "app is provisioned")
	return nil
}
