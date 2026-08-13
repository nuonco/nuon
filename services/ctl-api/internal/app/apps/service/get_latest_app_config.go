package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @ID						GetAppLatestConfig
// @Summary				get latest app config
// @Description.markdown	get_app_latest_config.md
// @Param					app_id	path	string	true	"app ID"
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
// @Router					/v1/apps/{app_id}/latest-config [get]
func (s *service) GetAppLatestConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	recurse := ctx.DefaultQuery("recurse", "false") == "true"

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	response, err := s.helpers.GetLatestActiveAppConfigBare(ctx, currentApp.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if recurse {
		response, err = s.helpers.GetFullAppConfig(ctx, response.ID, true)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get app config %s: %w", response.ID, err))
			return
		}
		if response == nil {
			ctx.Error(fmt.Errorf("no configs found for app: %w", gorm.ErrRecordNotFound))
			return
		}
	}

	ctx.JSON(http.StatusOK, response)
}
