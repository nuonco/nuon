package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
)

func (w *Workflows) pollChildrenDeprovisioned(ctx workflow.Context, appID string) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)
	for {
		currentApp, err := activities.AwaitGetByAppID(ctx, appID)
		if err != nil {
			w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
			return fmt.Errorf("unable to get app from database: %w", err)
		}

		installCnt := 0
		for _, install := range currentApp.Installs {
			// if an install was never attempted, it does not need to be polled
			if len(install.InstallSandboxRuns) < 1 {
				continue
			}

			if install.InstallSandboxRuns[0].Status != app.SandboxRunStatusAccessError {
				installCnt += 1
			}
		}

		if len(currentApp.Components) < 1 && installCnt < 1 {
			return nil
		}

		if workflow.Now(ctx).After(deadline) {
			err := fmt.Errorf("timeout waiting for installs and components to deprovision")
			w.updateStatus(ctx, appID, "error", err.Error())
			return err
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}

func (w *Workflows) deprovision(ctx workflow.Context, appID string, sandboxMode bool) error {
	w.updateStatus(ctx, appID, app.AppStatusActive, "polling for installs and components to be deprovisioned")
	if err := w.pollChildrenDeprovisioned(ctx, appID); err != nil {
		return err
	}

	// update status
	w.updateStatus(ctx, appID, app.AppStatusDeprovisioning, "deleting app resources")

	currentApp, err := activities.AwaitGetByAppID(ctx, appID)
	if err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are
	// migrated.
	if currentApp.Org.OrgType != app.OrgTypeV2 {
		if err := w.deprovisionLegacy(ctx, currentApp.OrgID, appID, sandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org deprovision: %w", err)
		}

		return nil
	}

	var repoProvisionResp executors.DeprovisionECRRepositoryResponse
	repoProvisionReq := &executors.DeprovisionECRRepositoryRequest{
		OrgID: currentApp.OrgID,
		AppID: appID,
	}
	if err := w.execChildWorkflow(ctx, appID, executors.DeprovisionECRRepositoryWorkflowName, sandboxMode, repoProvisionReq, &repoProvisionResp); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to provision ECR repository")
		return fmt.Errorf("unable to provision ECR repository: %w", err)
	}

	// update status with response
	if err := activities.AwaitDeleteByAppID(ctx, appID); err != nil {
		w.updateStatus(ctx, appID, app.AppStatusError, "unable to delete app")
		return fmt.Errorf("unable to delete app: %w", err)
	}

	return nil
}
