package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) deprovisionLegacy(ctx workflow.Context, orgID, appID string, sandboxMode bool) error {
	_, err := w.execDeprovisionWorkflow(ctx, sandboxMode, &appsv1.DeprovisionRequest{
		OrgId: orgID,
		AppId: appID,
	})
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to deprovision app")
		return fmt.Errorf("unable to deprovision app: %w", err)
	}

	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		AppID: appID,
	}); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to delete app")
		return fmt.Errorf("unable to delete app: %w", err)
	}

	return nil
}
