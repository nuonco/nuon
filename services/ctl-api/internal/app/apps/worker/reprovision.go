package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) reprovision(ctx workflow.Context, appID string, sandboxMode bool) error {
	if err := w.ensureOrg(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "org is unhealthy")
		return err
	}

	w.updateStatus(ctx, appID, app.AppStatusProvisioning, "reprovisioning app resources")

	currentApp, err := activities.AwaitGetByAppID(ctx, appID)
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are
	// migrated.
	if currentApp.Org.OrgType != app.OrgTypeV2 {
		if err := w.reprovisionLegacy(ctx, currentApp.OrgID, appID, sandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org reprovision: %w", err)
		}

		return nil
	}

	var repoProvisionResp executors.ProvisionECRRepositoryResponse
	repoProvisionReq := &executors.ProvisionECRRepositoryRequest{
		OrgID: currentApp.OrgID,
		AppID: appID,
	}
	if err := w.execChildWorkflow(ctx, appID, executors.ProvisionECRRepositoryWorkflowName, sandboxMode, repoProvisionReq, &repoProvisionResp); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to reprovision ECR repository")
		return fmt.Errorf("unable to reprovision ECR repository: %w", err)
	}

	w.updateStatus(ctx, appID, app.AppStatusActive, "app resources are reprovisioned")
	return nil
}
