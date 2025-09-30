package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetLatestAppBreakGlassConfig
// @Summary				get latest app break glass config
// @Description.markdown	get_latest_app_break_glass_config.md
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
// @Success				200	{object}	app.AppBreakGlassConfig
// @Router					/v1/apps/{app_id}/latest-app-break-glass-config [get]
func (s *service) GetLatestAppSecretsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getLatestAppSecretsConfig(ctx, currentApp.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get secrets config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getLatestAppSecretsConfig(ctx context.Context, appID string) (*app.AppSecretsConfig, error) {
	var appSecretsCfg app.AppSecretsConfig

	res := s.db.WithContext(ctx).
		Where(app.AppSecretsConfig{
			AppID: appID,
		}).
		Preload("Secrets").
		Order("created_at desc").
		Limit(1).
		First(&appSecretsCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app secrets config")
	}

	return &appSecretsCfg, nil
}
