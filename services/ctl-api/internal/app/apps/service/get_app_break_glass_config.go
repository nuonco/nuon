package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppBreakGlassConfig
// @Summary				get app break_glass config
// @Description.markdown	get_app_break_glass_config.md
// @Param					app_id	path	string	true	"app ID"
// @Param break_glass_config_id path string	true	"app break glass config ID"
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
// @Success				200	{object}	app.AppBreakGlassConfig
// @Router					/v1/apps/{app_id}/break-glass-configs/{break_glass_config_id} [get]
func (s *service) GetAppBreakGlassConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	appBreakGlassConfigID := ctx.Param("break_glass_config_id")

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getAppBreakGlassConfig(ctx, currentApp.ID, appBreakGlassConfigID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get break_glass config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getAppBreakGlassConfig(ctx context.Context, appID, appBreakGlassConfigID string) (*app.AppBreakGlassConfig, error) {
	var appBreakGlassCfg app.AppBreakGlassConfig

	res := s.db.WithContext(ctx).
		Where(app.AppBreakGlassConfig{
			AppID: appID,
			ID:    appBreakGlassConfigID,
		}).
		Preload("AWSIAMRoles").
		Preload("AWSIAMRoles.AppAWSIAMPolicyConfigs").
		Order("created_at desc").
		Limit(1).
		First(&appBreakGlassCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app break_glass config")
	}

	return &appBreakGlassCfg, nil
}
