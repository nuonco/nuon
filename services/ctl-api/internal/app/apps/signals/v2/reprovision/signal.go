package reprovision

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-reprovision"

type Signal struct {
	AppID string `json:"app_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppID == "" {
		return errors.New("app_id is required")
	}

	// Validate app exists
	_, err := activities.AwaitGetByAppID(ctx, s.AppID)
	if err != nil {
		return errors.Wrap(err, "app not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Ensure org is healthy before reprovisioning
	currentApp, err := activities.AwaitGetByAppID(ctx, s.AppID)
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			AppID:             s.AppID,
			Status:            app.AppStatusError,
			StatusDescription: "unable to get app from database",
		}); updateErr != nil {
			l.Error("failed to update app status", zap.Error(updateErr))
		}
		return errors.Wrap(err, "unable to get app from database")
	}

	// TODO: Org health check disabled - orgsactivities.AwaitGetOrgStatus not implemented
	// Check org health
	// orgStatus, err := orgsactivities.AwaitGetOrgStatus(ctx, currentApp.OrgID)
	// if err != nil {
	// 	if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
	// 		AppID:             s.AppID,
	// 		Status:            app.AppStatusError,
	// 		StatusDescription: "unable to check org health",
	// 	}); updateErr != nil {
	// 		l.Error("failed to update app status", zap.Error(updateErr))
	// 	}
	// 	return errors.Wrap(err, "unable to get org status")
	// }
	//
	// if !orgStatus.IsHealthy() {
	// 	if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
	// 		AppID:             s.AppID,
	// 		Status:            app.AppStatusError,
	// 		StatusDescription: "org is unhealthy",
	// 	}); updateErr != nil {
	// 		l.Error("failed to update app status", zap.Error(updateErr))
	// 	}
	// 	return errors.Errorf("org is unhealthy: %s", orgStatus)
	// }

	// Update status to provisioning
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		AppID:             s.AppID,
		Status:            app.AppStatusProvisioning,
		StatusDescription: "reprovisioning app resources",
	}); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	// Provision ECR repository (only for default org type)
	var repoResp *ecrrepository.ProvisionECRRepositoryResponse
	if currentApp.Org.OrgType == app.OrgTypeDefault {
		repoResp, err = ecrrepository.AwaitProvisionECRRepository(ctx, &ecrrepository.ProvisionECRRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: s.AppID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to provision ECR repository")
		}
	} else {
		l.Info("skipping reprovision ecr",
			zap.String("app_id", currentApp.ID),
			zap.String("app_name", currentApp.Name),
			zap.Any("org_type", currentApp.Org.OrgType),
			zap.String("org_id", currentApp.Org.ID),
			zap.String("org_name", currentApp.Org.Name))
	}

	// Create app repository record
	if _, err := activities.AwaitCreateAppRepository(ctx, &activities.CreateAppRepositoryRequest{
		AppID:          s.AppID,
		CreateResponse: repoResp,
	}); err != nil {
		return errors.Wrap(err, "unable to create app repository")
	}

	// Update status to active
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		AppID:             s.AppID,
		Status:            app.AppStatusActive,
		StatusDescription: "app resources are reprovisioned",
	}); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	return nil
}
