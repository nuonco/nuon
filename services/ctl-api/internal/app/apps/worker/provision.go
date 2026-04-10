package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
)

// @temporal-gen-v2 workflow
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

	var repoResp *ecrrepository.ProvisionECRRepositoryResponse
	switch {
	case currentApp.Org.OrgType != app.OrgTypeDefault:
		repoResp = generics.GetFakeObj[*ecrrepository.ProvisionECRRepositoryResponse]()
		l.Info("skipping provision ecr",
			zap.String("app_id", currentApp.ID),
			zap.Any("org_type", currentApp.Org.OrgType))
	case w.cfg.IsGCP():
		// GCP uses GAR — the management repository is shared, apps use org/app path prefix.
		garURL := w.cfg.ManagementGARRepositoryURL
		repoResp = &ecrrepository.ProvisionECRRepositoryResponse{
			RepositoryName: fmt.Sprintf("%s/%s", currentApp.OrgID, sreq.ID),
			RepositoryURI:  fmt.Sprintf("%s/%s/%s", garURL, currentApp.OrgID, sreq.ID),
			Region:         w.cfg.AppRegion,
		}
		l.Info("using GAR repository",
			zap.String("app_id", currentApp.ID),
			zap.String("repository_uri", repoResp.RepositoryURI))
	case w.cfg.IsAzure():
		repoResp = w.acrRepositoryResponse(currentApp.OrgID, sreq.ID)
		l.Info("using ACR repository",
			zap.String("app_id", currentApp.ID),
			zap.String("repository_uri", repoResp.RepositoryURI))
	default:
		repoResp, err = ecrrepository.AwaitProvisionECRRepository(ctx, &ecrrepository.ProvisionECRRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: sreq.ID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to provision ECR repository")
		}
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
