package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestartAppRequest struct{}

//	@BasePath	/v1/apps
//
// Restart an app's event loop
//
//	@Summary	restart an apps event loop
//	@Schemes
//	@Description	restart app event loop
//	@Param			app_id	path	string					true	"app ID"
//	@Param			req			body	RestartAppRequest	true	"Input"
//	@Tags			apps/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/apps/{app_id}/restart [POST]
func (s *service) RestartApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req RestartAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	app, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	s.hooks.Restart(ctx, app.ID)
	ctx.JSON(http.StatusOK, true)
}
