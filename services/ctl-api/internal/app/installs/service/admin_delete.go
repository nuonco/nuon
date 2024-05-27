package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminDeleteInstallRequest struct{}

// @ID AdminDeleteInstall
// @Summary	delete an install
// @Description.markdown delete_install.md
// @Tags			installs/admin
// @Accept			json
// @Param			req			body	AdminDeleteInstallRequest	true	"Input"
// @Param			install_id	path	string						true	"install id"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/installs/{install_id}/admin-delete [POST]
func (s *service) AdminDeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationTeardownComponents,
	})
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	ctx.JSON(http.StatusOK, true)
}
