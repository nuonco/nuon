package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
	if err := w.ensureOrg(ctx, sreq.ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.AppStatusError, "org is unhealthy")
		return err
	}

	w.updateStatus(ctx, sreq.ID, app.AppStatusProvisioning, "provisioning app resources")

	currentApp, err := activities.AwaitGetByAppID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are
	// migrated.
	if currentApp.Org.OrgType == app.OrgTypeLegacy {
		if err := w.provisionLegacy(ctx, currentApp.OrgID, sreq.ID, sreq.SandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org provision: %w", err)
		}

		return nil
	}

	if currentApp.Org.OrgType == app.OrgTypeDefault {
		var repoProvisionResp executors.ProvisionECRRepositoryResponse
		repoProvisionReq := &executors.ProvisionECRRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: sreq.ID,
		}
		if err := w.execChildWorkflow(ctx, sreq.ID, executors.ProvisionECRRepositoryWorkflowName, sreq.SandboxMode, repoProvisionReq, &repoProvisionResp); err != nil {
			w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to provision ECR repository")
			return fmt.Errorf("unable to provision ECR repository: %w", err)
		}
	} else {
		l.Info("skipping provision ecr",
			zap.String("app_id", currentApp.ID),
			zap.String("app_name", currentApp.Name),
			zap.Any("org_type", currentApp.Org.OrgType),
			zap.String("org_id", currentApp.Org.ID),
			zap.String("org_name", currentApp.Org.Name))
	}

	// update status with response
	w.updateStatus(ctx, sreq.ID, app.AppStatusActive, "app resources are provisioned")
	return nil
}
