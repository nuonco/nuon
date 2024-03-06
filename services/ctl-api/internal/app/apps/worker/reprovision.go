package worker

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) reprovision(ctx workflow.Context, appID string, dryRun bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, StatusError, "org is unhealthy")
		return err
	}

	w.updateStatus(ctx, appID, StatusProvisioning, "reprovisioning app resources")

	var app app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		AppID: appID,
	}, &app); err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to get app from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	_, err := w.execProvisionWorkflow(ctx, dryRun, &appsv1.ProvisionRequest{
		OrgId: app.OrgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, StatusError, "unable to reprovision app resources")
		return fmt.Errorf("unable to provision app: %w", err)
	}

	w.updateStatus(ctx, appID, StatusActive, "app resources are provisioned")
	return nil
}
