package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

//	@BasePath	/v1/apps
//
// Get app input configs
//
//	@Summary	get app input configs
//	@Schemes
//	@Description	get app input configs
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
//	@Success		200				{object}	[]app.AppInputConfig
//	@Router			/v1/apps/{app_id}/input-configs [get]
func (s *service) GetAppInputConfigs(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, app.AppInputConfigs)
}
