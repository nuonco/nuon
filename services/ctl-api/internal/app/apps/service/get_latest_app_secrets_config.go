package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetLatestAppSecretsConfig
// @Summary				get latest app secrets config
// @Description.markdown	get_latest_app_secrets_config.md
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
// @Success				200	{object}	app.AppSecretsConfig
// @Router					/v1/apps/{app_id}/latest-app-secrets-config [get]
func (s *service) GetLatestAppBreakGlassConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.getLatestAppBreakGlassConfig(ctx, currentApp.ID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get secrets config"))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}

func (s *service) getLatestAppBreakGlassConfig(ctx context.Context, appID string) (*app.AppBreakGlassConfig, error) {
	var appBreakGlassCfg app.AppBreakGlassConfig

	res := s.db.WithContext(ctx).
		Where(app.AppBreakGlassConfig{
			AppID: appID,
		}).
		Preload("AWSIAMRoles").
		Order("created_at desc").
		Limit(1).
		First(&appBreakGlassCfg)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app secrets config")
	}

	return &appBreakGlassCfg, nil
}
