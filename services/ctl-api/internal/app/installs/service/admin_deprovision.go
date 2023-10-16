package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDeprovisionInstallRequest struct{}

// Deprovision an install
//
//	@Summary	deprovision an install, but keep it in the database
//
//	@Schemes
//
//	@Description	deprovision an install
//
//
//	@Tags			installs/admin
//	@Accept			json
//	@Param			req			body	AdminDeprovisionInstallRequest	true	"Input"
//
//	@Param			install_id	path	string							true	"org ID for your current org"
//
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/installs/{install_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	s.hooks.Deprovisioned(ctx, installID)
	ctx.JSON(http.StatusOK, true)
}
