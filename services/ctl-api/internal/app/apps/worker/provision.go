package worker

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, appID string, dryRun bool) error {
	if err := w.pollDependencies(ctx, appID); err != nil {
		return fmt.Errorf("unable to poll org for app: %w", err)
	}

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &currentApp); err != nil {
		return fmt.Errorf("unable to get app: %w", err)
	}

	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		AppID:             appID,
		Status:            "provisioning",
		StatusDescription: "creating app resources - this should just take a few seconds",
	}); err != nil {
		return fmt.Errorf("unable to update app status: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &appsv1.ProvisionRequest{
		OrgId: currentApp.OrgID,
		AppId: appID,
	})
	if err != nil {
		return fmt.Errorf("unable to provision app: %w", err)
	}

	// update status with response
	if err := w.defaultExecErrorActivity(ctx, w.acts.UpdateStatus, activities.UpdateStatusRequest{
		AppID:             appID,
		Status:            "active",
		StatusDescription: "active",
	}); err != nil {
		return fmt.Errorf("unable to update app status: %w", err)
	}
	return nil
}
