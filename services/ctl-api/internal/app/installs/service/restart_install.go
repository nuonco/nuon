package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestartInstallRequest struct{}

//	@BasePath	/v1/installs
//
// Restart an install's event loop
//
//	@Summary	restart an installs event loop
//	@Schemes
//	@Description	restart install event loop
//	@Param			install_id	path	string					true	"install ID"
//	@Param			req			body	RestartInstallRequest	true	"Input"
//	@Tags			installs/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/installs/{install_id}/admin-restart [POST]
func (s *service) RestartInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req RestartInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.hooks.Restart(ctx, install.ID, install.App.Org.SandboxMode)
	ctx.JSON(http.StatusOK, true)
}
