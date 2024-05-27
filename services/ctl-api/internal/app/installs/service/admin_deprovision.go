package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminDeprovisionInstallRequest struct{}

// @ID AdminDeprovisionInstall
// @Summary	deprovision an install, but keep it in the database
// @Description.markdown deprovision_install.md
// @Tags			installs/admin
// @Accept			json
// @Param			req			body	AdminDeprovisionInstallRequest	true	"Input"
// @Param	install_id	path	string	true	"org ID for your current org"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/installs/{install_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationDeprovision,
	})
	ctx.JSON(http.StatusOK, true)
}
