package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	appsv1 "github.com/nuonco/nuon/pkg/types/workflows/apps/v1"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (w *Workflows) provisionLegacy(ctx workflow.Context, orgID, appID string, sandboxMode bool) error {
	_, err := w.execProvisionWorkflow(ctx, sandboxMode, &appsv1.ProvisionRequest{
		OrgId: orgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to provision app")
		return fmt.Errorf("unable to provision app: %w", err)
	}
	w.updateStatus(ctx, appID, app.AppStatusActive, "app resources are provisioned")
	return nil
}
