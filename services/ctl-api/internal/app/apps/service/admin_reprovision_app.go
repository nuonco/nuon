package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

type ReprovisionAppRequest struct{}

//	@ID						AdminReprovisionApp
//	@Summary				reprovision an app
//	@Description.markdown	reprovision_app.md
//	@Param					app_id	path	string	true	"app ID for your current app"
//	@Tags					apps/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	ReprovisionAppRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/apps/{app_id}/admin-reprovision [POST]
func (s *service) AdminReprovisionApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	_, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, appID, &signals.Signal{
		Type: signals.OperationReprovision,
	})
	ctx.JSON(http.StatusOK, true)
}
