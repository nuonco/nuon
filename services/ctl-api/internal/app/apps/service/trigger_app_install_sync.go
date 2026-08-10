package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/syncinstalls"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						TriggerAppInstallSync
// @Summary				trigger app-level install config sync
// @Description			Triggers a sync of all install configs for the app from the configured git source.
// @Tags					apps
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					app_id	path	string	true	"app ID"
// @Success				202	{object}	app.AppInstallConfigSync
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/install-syncs [post]
func (s *service) TriggerAppInstallSync(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")

	var a app.App
	if err := s.db.WithContext(ctx).
		Where(app.App{ID: appID, OrgID: org.ID}).
		First(&a).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	var latestConfig app.AppInstallsConfig
	if err := s.db.WithContext(ctx).
		Where(app.AppInstallsConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		ctx.Error(fmt.Errorf("no installs config found for app - configure via installs.toml or the dashboard"))
		return
	}

	if err := s.helpers.EnsureAppQueue(ctx, appID); err != nil {
		ctx.Error(fmt.Errorf("unable to ensure app queues: %w", err))
		return
	}

	syncRecord := app.AppInstallConfigSync{
		AppID:       appID,
		TriggeredBy: "manual",
		Status: app.CompositeStatus{
			Status: app.StatusQueued,
		},
	}
	if err := s.db.WithContext(ctx).Create(&syncRecord).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create install config sync: %w", err))
		return
	}

	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, "apps", appshelpers.AppInstallSyncsQueueName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find install-syncs queue for app: %w", err))
		return
	}

	enqueueResp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   syncRecord.ID,
		OwnerType: "app_install_config_syncs",
		Signal: &syncinstalls.Signal{
			AppID:                  appID,
			AppInstallConfigSyncID: syncRecord.ID,
			TriggeredBy:            "manual",
			FallbackCreatedByID:    account.ID,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue sync signal: %w", err))
		return
	}

	s.db.WithContext(ctx).Model(&syncRecord).Updates(map[string]any{
		"queue_signal_id": enqueueResp.ID,
		"queue_id":        queue.ID,
	})
	syncRecord.QueueSignalID = enqueueResp.ID
	syncRecord.QueueID = queue.ID

	ctx.JSON(http.StatusAccepted, syncRecord)
}
