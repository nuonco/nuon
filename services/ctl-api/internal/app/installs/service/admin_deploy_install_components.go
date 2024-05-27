package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminDeployInstallComponentsRequest struct{}

// @ID AdminDeployInstallComponents
// @Summary	deploy all components on an install
// @Description.markdown deploy_install_components.md
// @Param			install_id	path	string					true	"install ID"
// @Param			req			body	AdminDeployInstallComponentsRequest	true	"Input"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/installs/{install_id}/admin-deploy-components [POST]
func (s *service) AdminDeployInstallComponents(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminDeployInstallComponentsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationDeployComponents,
	})
	ctx.JSON(http.StatusOK, true)
}
