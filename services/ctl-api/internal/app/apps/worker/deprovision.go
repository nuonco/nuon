package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
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

			if install.InstallSandboxRuns[0].Status != app.SandboxRunStatusAccessError &&
				install.InstallSandboxRuns[0].Status != app.SandboxRunStatusDeprovisioned {
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

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) Deprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
	w.updateStatus(ctx, sreq.ID, app.AppStatusActive, "polling for installs and components to be deprovisioned")
	if err := w.pollChildrenDeprovisioned(ctx, sreq.ID); err != nil {
		return err
	}

	// update status
	w.updateStatus(ctx, sreq.ID, app.AppStatusDeprovisioning, "deleting app resources")

	currentApp, err := activities.AwaitGetByAppID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to get app from database")
		return fmt.Errorf("unable to get app from database: %w", err)
	}

	// NOTE(jm): this will be removed once the runner is in prod and all orgs are
	// migrated.
	if currentApp.Org.OrgType == app.OrgTypeLegacy {
		if err := w.deprovisionLegacy(ctx, currentApp.OrgID, sreq.ID, sreq.SandboxMode); err != nil {
			return fmt.Errorf("unable to perform legacy org deprovision: %w", err)
		}

		return nil
	}

	if currentApp.Org.OrgType == app.OrgTypeDefault {
		var repoProvisionResp executors.DeprovisionECRRepositoryResponse
		repoProvisionReq := &executors.DeprovisionECRRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: sreq.ID,
		}
		if err := w.execChildWorkflow(ctx, sreq.ID, executors.DeprovisionECRRepositoryWorkflowName, sreq.SandboxMode, repoProvisionReq, &repoProvisionResp); err != nil {
			w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to provision ECR repository")
			return fmt.Errorf("unable to provision ECR repository: %w", err)
		}
	} else {
		l.Info("skipping deprovision ecr",
			zap.String("app_id", currentApp.ID),
			zap.String("app_name", currentApp.Name),
			zap.Any("org_type", currentApp.Org.OrgType),
			zap.String("org_id", currentApp.Org.ID),
			zap.String("org_name", currentApp.Org.Name))
	}

	// update status with response
	if err := activities.AwaitDeleteByAppID(ctx, sreq.ID); err != nil {
		w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to delete app")
		return fmt.Errorf("unable to delete app: %w", err)
	}

	return nil
}
