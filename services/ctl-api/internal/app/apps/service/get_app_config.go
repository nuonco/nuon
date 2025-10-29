package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppConflgV2
// @Summary				get an app config
// @Description.markdown	get_app_config.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Param recurse query bool false "load all children configs" Default(false)
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
// @Success				200	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/configs/{app_config_id} [get]
func (s *service) GetAppConfigV2(ctx *gin.Context) {
	s.GetAppConfig(ctx)
}

// @ID						GetAppConfig
// @Summary				get an app config
// @Description.markdown	get_app_config.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Param recurse query bool false "load all children configs" Default(false)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/config/{app_config_id} [get]
func (s *service) GetAppConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	appConfigID := ctx.Param("app_config_id")

	recurse := ctx.DefaultQuery("recurse", "false") == "true"

	var appConfig *app.AppConfig
	if recurse {
		appConfig, err = s.helpers.GetFullAppConfig(ctx, appConfigID, true)
	} else {
		appConfig, err = s.getAppConfig(ctx, org.ID, appID, appConfigID)
	}

	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, appConfig)
}

func (s *service) getAppConfig(ctx context.Context, orgID, appID, appCfgID string) (*app.AppConfig, error) {
	appCfg := app.AppConfig{}
	res := s.db.WithContext(ctx).
		Where(app.AppConfig{
			OrgID: orgID,
			AppID: appID,
		}).
		First(&appCfg, "id = ?", appCfgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	return &appCfg, nil
}
