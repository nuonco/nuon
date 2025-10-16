package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						GetRunnerAppConfig
// @Summary				get an app config in the runner context
// @Description.markdown	get_app_config.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Tags apps/runner
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
// @Router					/v1/apps/{app_id}/config/{app_config_id} [get]
func (s *service) GetRunnerAppConfig(ctx *gin.Context) {
	appConfigID := ctx.Param("app_config_id")

	appConfig, err := s.helpers.GetFullAppConfig(ctx, appConfigID, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, appConfig)
}
