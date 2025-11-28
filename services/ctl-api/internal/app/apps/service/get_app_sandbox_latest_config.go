package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

// @ID						GetAppSandboxLatestConfig
// @Summary				get latest app sandbox config
// @Description.markdown	get_app_sandbox_latest_config.md
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
// @Success				200	{object}	app.AppSandboxConfig
// @Router					/v1/apps/{app_id}/sandbox-latest-config [get]
func (s *service) GetAppSandboxLatestConfig(ctx *gin.Context) {
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

	if len(app.AppSandboxConfigs) < 1 {
		ctx.Error(fmt.Errorf("no app sandbox configs found for app: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, app.AppSandboxConfigs[0])
}
