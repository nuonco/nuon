package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	appdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/deprovision"
	componentssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	componentdelete "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/delete"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						DeleteApp
// @Summary				delete an app
// @Description.markdown	delete_app.md
// @Param					app_id	path	string	true	"app ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id} [DELETE]
func (s *service) DeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	// Check if there are any active installs for the app, if so, do not allow deletion.
	{
		installs, err := s.getAppInstalls(ctx, appID)
		if err != nil {
			ctx.Error(fmt.Errorf("error fetching app installs: %w", err))
			return
		}

		activeInstalls := make([]string, 0)
		for _, ins := range installs {
			// if an install was never attempted, it does not need to be accounted for
			if len(ins.InstallSandboxRuns) < 1 {
				continue
			}

			if ins.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError ||
				ins.InstallSandboxRuns[0].Status == app.SandboxRunStatusDeprovisioned ||
				ins.InstallSandboxRuns[0].Status == app.SandboxRunStatusDeprovisioning {
				continue
			}

			activeInstalls = append(activeInstalls, ins.ID)
		}
		if len(activeInstalls) > 0 {
			ctx.Status(http.StatusBadRequest)
			ctx.Error(fmt.Errorf("app has %d active install(s), please deprovision them first", len(activeInstalls)))
			return
		}
	}

	appCfg, cfgErr := s.helpers.GetAppLatestConfig(ctx, appID)
	if cfgErr != nil && !errors.Is(cfgErr, gorm.ErrRecordNotFound) {
		ctx.Error(cfgErr)
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}

	if cfgErr == nil {
		// Check if there are any active components for the app, if so, do not proceed for deletion.
		if len(appCfg.ComponentIDs) > 0 {
			ctx.Error(fmt.Errorf("app has %d active component(s) in it's latest config, please remove them first", len(appCfg.ComponentIDs)))
			return
		}

		// Trigger deletion for all components associated with the app in reverse order of their dependencies.
		{
			// Get full app config to include all components including missing components in the latest config.
			appComponents, err := s.helpers.GetAppComponents(ctx, appID)
			if err != nil {
				ctx.Error(err)
				return
			}

			for _, comp := range appComponents {
				if err := s.helpers.DeleteAppComponent(ctx, comp.ID); err != nil {
					ctx.Error(err)
					return
				}

				if useQueues {
					q, err := s.queueClient.GetQueueByOwner(ctx, comp.ID, "components")
					if err != nil {
						ctx.Error(fmt.Errorf("unable to get component queue: %w", err))
						return
					}
					if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
						QueueID:   q.ID,
						OwnerID:   comp.ID,
						OwnerType: "components",
						Signal:    &componentdelete.Signal{ComponentID: comp.ID},
					}); err != nil {
						ctx.Error(fmt.Errorf("unable to enqueue component delete signal: %w", err))
						return
					}
				} else {
					s.evClient.Send(ctx, comp.ID, &componentssignals.Signal{
						Type: componentssignals.OperationDelete,
					})
				}
			}
		}
	}

	if err := s.deleteApp(ctx, appID); err != nil {
		ctx.Error(err)
		return
	}

	if useQueues {
		q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get app queue: %w", err))
			return
		}
		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   appID,
			OwnerType: "apps",
			Signal:    &appdeprovision.Signal{AppID: appID},
		}); err != nil {
			ctx.Error(fmt.Errorf("unable to enqueue app deprovision signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, appID, &signals.Signal{
			Type: signals.OperationDeprovision,
		})
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) deleteApp(ctx context.Context, appID string) error {
	currentApp := app.App{
		ID: appID,
	}

	res := s.db.WithContext(ctx).Model(&currentApp).Updates(app.App{
		Status:            app.AppStatusDeleteQueued,
		StatusDescription: "delete has been queued and waiting",
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update app: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("app not found %s: %w", appID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *service) getAppInstalls(ctx context.Context, appID string) ([]app.Install, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Installs").
		Preload("Installs.InstallSandboxRuns").
		Where("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return app.Installs, nil
}
