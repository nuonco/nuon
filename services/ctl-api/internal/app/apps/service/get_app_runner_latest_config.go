package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

// @ID						GetAppRunnerLatestConfig
// @Summary				get latest app runner config
// @Description.markdown	get_app_runner_latest_config.md
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
// @Success				200	{object}	app.AppRunnerConfig
// @Router					/v1/apps/{app_id}/runner-latest-config [get]
func (s *service) GetAppRunnerLatestConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
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

	if len(app.AppRunnerConfigs) < 1 {
		ctx.Error(fmt.Errorf("no app runner configs found for app: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, app.AppRunnerConfigs[0])
}
