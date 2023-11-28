package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDeleteAppRequest struct{}

// Delete an app
//
//	@Summary	delete an app
//
//	@Schemes
//
//	@Description	delete an app
//
//	@Tags			apps/admin
//	@Accept			json
//	@Param			req			body	AdminDeleteAppRequest	true	"Input"
//	@Param			app_id	path	string						true	"app id"
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/apps/{app_id}/admin-delete [POST]
func (s *service) AdminDeleteApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	s.hooks.Deleted(ctx, appID)
	ctx.JSON(http.StatusOK, true)
}
