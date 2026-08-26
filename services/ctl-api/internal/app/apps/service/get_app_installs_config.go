package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppInstallsConfig
// @Summary				get latest installs config for an app
// @Description			Returns the latest installs config (git source for install config files).
// @Tags					apps
// @Produce				json
// @Security				APIKey && OrgID
// @Param					app_id	path	string	true	"app ID"
// @Success				200	{object}	app.AppInstallsConfig
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/installs-configs [get]
func (s *service) GetAppInstallsConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")

	var cfg app.AppInstallsConfig
	if err := s.db.WithContext(ctx).
		Where(app.AppInstallsConfig{AppID: appID, OrgID: org.ID}).
		Order("created_at DESC").
		First(&cfg).Error; err != nil {
		ctx.Error(fmt.Errorf("no installs config found for app: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cfg)
}
