package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReprovisionInstallRequest struct{}

// @ID AdminReprovisionInstall
// @Description.markdown reprovision_install.md
// @Param			install_id	path	string	true	"install ID for your current install"
// @Tags			installs/admin
// @Accept			json
// @Param			req	body	ReprovisionInstallRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/installs/{install_id}/admin-reprovision [POST]
func (s *service) ReprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Reprovision(ctx, installID)
	ctx.JSON(http.StatusOK, true)
}
