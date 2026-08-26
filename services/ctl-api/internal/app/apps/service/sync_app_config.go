package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/appconfigsync"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						SyncAppConfig
// @Description.markdown	sync_app_config.md
// @Tags					apps
// @Accept					json
// @Produce				json
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/configs/{config_id}/sync [post]
func (s *service) SyncAppConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	var appConfig app.AppConfig
	res := s.db.WithContext(ctx).
		Where(app.AppConfig{OrgID: org.ID, AppID: appID}).
		First(&appConfig, "id = ?", configID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app config: %w", res.Error))
		return
	}

	if appConfig.Status == app.AppConfigStatusSyncing {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("app config %s is already syncing", configID),
			Description: "this app config is already syncing",
			Code:        "app_config_already_syncing",
		})
		return
	}

	if appConfig.IntermediateConfig == nil || !appConfig.IntermediateConfig.IsSet() {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("app config %s has no intermediate config", configID),
			Description: "this app config was not created with an intermediate config, so there is nothing to sync",
			Code:        "app_config_missing_intermediate_config",
		})
		return
	}

	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app queue: %w", err))
		return
	}

	accountID := ""
	if acct, err := cctx.AccountFromGinContext(ctx); err == nil {
		accountID = acct.ID
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &appconfigsync.Signal{
			AppID:       appID,
			AppConfigID: configID,
			AccountID:   accountID,
		},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue app config sync signal: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, &appConfig)
}
