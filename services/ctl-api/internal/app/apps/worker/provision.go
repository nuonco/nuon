package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/ecrrepository"
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

	var repoResp *ecrrepository.CreateRepositoryResponse
	if currentApp.Org.OrgType == app.OrgTypeDefault {
		repoResp, err = ecrrepository.AwaitCreateRepository(ctx, &ecrrepository.CreateRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: sreq.ID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to provision ECR repository")
		}
	} else {
		repoResp = generics.GetFakeObj[*ecrrepository.CreateRepositoryResponse]()
		l.Info("skipping provision ecr",
			zap.String("app_id", currentApp.ID),
			zap.String("app_name", currentApp.Name),
			zap.Any("org_type", currentApp.Org.OrgType),
			zap.String("org_id", currentApp.Org.ID),
			zap.String("org_name", currentApp.Org.Name))
	}

	if _, err := activities.AwaitCreateAppRepository(ctx, &activities.CreateAppRepositoryRequest{
		AppID:          sreq.ID,
		CreateResponse: repoResp,
	}); err != nil {
		return errors.Wrap(err, "unable to create app repository")
	}

	// update status with response
	w.updateStatus(ctx, sreq.ID, app.AppStatusActive, "app resources are provisioned")
	return nil
}
