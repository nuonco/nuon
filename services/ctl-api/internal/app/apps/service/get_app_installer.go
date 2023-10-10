package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@BasePath	/v1/apps
//
// Create get an app
//
//	@Summary	get an app installer
//	@Schemes
//	@Description	get an app installer
//	@Param			installer_id	path	string	true	"installer ID"
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
//	@Success		200				{object}	app.AppInstaller
//	@Router			/v1/installers/{installer_id} [get]
func (s *service) GetAppInstaller(ctx *gin.Context) {
	installerID := ctx.Param("installer_id")
	app, err := s.getAppInstaller(ctx, installerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app installer %s: %w", installerID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}
