package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"gorm.io/gorm"
)

// @ID GetAppLatestConfig
// @Summary	get latest app config
// @Description.markdown	get_app_latest_config.md
// @Param			app_id	path	string	true	"app ID"
// @Tags			apps
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.AppConfig
// @Router			/v1/apps/{app_id}/latest-config [get]
func (s *service) GetAppLatestConfig(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	app, err := s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	if len(app.AppConfigs) < 1 {
		ctx.Error(fmt.Errorf("no configs found for app: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, app.AppConfigs[0])
}
