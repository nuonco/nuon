package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteAppInstallsConfig
// @Summary				soft-delete an installs config
// @Description			Soft-deletes an installs config record. The next latest record becomes active.
// @Tags					apps
// @Produce				json
// @Security				APIKey && OrgID
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"config ID"
// @Success				200	{object}	map[string]string
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/installs-configs/{config_id} [delete]
func (s *service) DeleteAppInstallsConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	result := s.db.WithContext(ctx).
		Where(app.AppInstallsConfig{ID: configID, AppID: appID, OrgID: org.ID}).
		Delete(&app.AppInstallsConfig{})

	if result.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete installs config: %w", result.Error))
		return
	}

	if result.RowsAffected == 0 {
		ctx.Error(fmt.Errorf("installs config not found"))
		return
	}

	ctx.JSON(http.StatusOK, map[string]string{
		"status": "deleted",
	})
}
