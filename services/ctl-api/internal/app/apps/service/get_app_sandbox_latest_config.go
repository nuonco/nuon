package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"gorm.io/gorm"
)

//	@BasePath	/v1/apps
//
// Get latest app sandbox config
//
//	@Summary	get latest app sandbox config
//	@Schemes
//	@Description	get latest app sandbox config
//
// @Param			app_id	path	string	true	"app ID"
//
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.AppSandboxConfig
//	@Router			/v1/apps/{app_id}/sandbox-latest-config [get]
func (s *service) GetAppSandboxLatestConfig(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
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
