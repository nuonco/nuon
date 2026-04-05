package deprovision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-deprovision"

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

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

	// Update status - polling for children to deprovision
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		AppID:             s.AppID,
		Status:            app.AppStatusActive,
		StatusDescription: "polling for installs to be deprovisioned",
	}); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	// Poll until all children are deprovisioned
	if err := s.pollChildrenDeprovisioned(ctx); err != nil {
		return err
	}

	// Update status to deprovisioning
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		AppID:             s.AppID,
		Status:            app.AppStatusDeprovisioning,
		StatusDescription: "deleting app resources",
	}); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	// Get current app
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

	// Deprovision ECR repository (only for default org type)
	if currentApp.Org.OrgType == app.OrgTypeDefault {
		repoDeprovisionReq := &ecrrepository.DeprovisionECRRepositoryRequest{
			OrgID: currentApp.OrgID,
			AppID: s.AppID,
		}
		_, err := ecrrepository.AwaitDeprovisionECRRepository(ctx, repoDeprovisionReq)
		if err != nil {
			return errors.Wrap(err, "unable to deprovision ECR repository")
		}
	} else {
		l.Info("skipping deprovision ecr",
			zap.String("app_id", currentApp.ID),
			zap.String("app_name", currentApp.Name),
			zap.Any("org_type", currentApp.Org.OrgType),
			zap.String("org_id", currentApp.Org.ID),
			zap.String("org_name", currentApp.Org.Name))
	}

	// Delete the app
	if err := activities.AwaitDeleteByAppID(ctx, s.AppID); err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			AppID:             s.AppID,
			Status:            app.AppStatusError,
			StatusDescription: "unable to delete app",
		}); updateErr != nil {
			l.Error("failed to update app status", zap.Error(updateErr))
		}
		return errors.Wrap(err, "unable to delete app")
	}

	return nil
}

func (s *Signal) pollChildrenDeprovisioned(ctx workflow.Context) error {
	deadline := workflow.Now(ctx).Add(time.Minute * 60)

	for {
		currentApp, err := activities.AwaitGetByAppID(ctx, s.AppID)
		if err != nil {
			if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
				AppID:             s.AppID,
				Status:            app.AppStatusError,
				StatusDescription: "unable to get app from database",
			}); updateErr != nil {
				workflow.GetLogger(ctx).Error("failed to update app status", updateErr)
			}
			return fmt.Errorf("unable to get app from database: %w", err)
		}

		// Count installs that need to be deprovisioned
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

		// If no installs remaining, we're done.
		// Components are cascade-deleted when the app is deleted.
		if installCnt < 1 {
			return nil
		}

		// Check timeout
		if workflow.Now(ctx).After(deadline) {
			err := fmt.Errorf("timeout waiting for installs to deprovision")
			if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
				AppID:             s.AppID,
				Status:            "error",
				StatusDescription: err.Error(),
			}); updateErr != nil {
				workflow.GetLogger(ctx).Error("failed to update app status", updateErr)
			}
			return err
		}

		// Sleep and poll again
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
