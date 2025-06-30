package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppConfigSyncRun
// @Summary				get the work record of syncing an app config
// @Description.markdown	get_app_config_sync_run.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_sync_run_id	path	string	true	"app config sync run ID"
// @Tags					apps/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppConfigSyncRun
// @Router					/v1/apps/{app_id}/config-sync-run/{app_config_sync_run_id} [get]
func (s *service) GetAppConfigSyncRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	acsrID := ctx.Param("app_config_sync_run_id")

	acsr, err := s.getAppConfigSyncRun(ctx, org.ID, appID, acsrID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, acsr)
}

func (s *service) getAppConfigSyncRun(ctx context.Context, orgID, appID, acsrID string) (*app.AppConfigSyncRun, error) {
	acsr := new(app.AppConfigSyncRun)
	res := s.db.WithContext(ctx).
		Where(app.AppConfigSyncRun{
			OrgID: orgID,
			AppID: appID,
			ID:    acsrID,
		}).
		First(acsr, "id = ?", acsrID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config sync run: %w", res.Error)
	}

	return acsr, nil
}

