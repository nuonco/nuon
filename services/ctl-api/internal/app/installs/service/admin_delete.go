package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDeleteInstallRequest struct{}

// Delete an install
//
//	@Summary deprovision an install, but keep it in the database
//
// @Schemes
//
//	@Description	deprovision an install
//
// @Param			install_id	path	string	true	"org ID for your current org"
//
//	@Tags			installs/admin
//	@Accept			json
//	@Param			req	body	AdminDeleteInstallRequest	true	"Input"
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/installs/{install_id}/admin-deprovision [POST]
func (s *service) AdminDeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	s.hooks.Deleted(ctx, installID)
	ctx.JSON(http.StatusOK, true)
}
