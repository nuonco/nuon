package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminReprovisionInstallRunnerRequest struct{}

//	@ID						AdminReprovisionInstallRunner
//	@Description.markdown	reprovision_install_runner.md
//	@Param					install_id	path	string	true	"install ID for your current install"
//	@Tags					installs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	ReprovisionInstallRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/installs/{install_id}/admin-reprovision-runner [POST]
func (s *service) AdminReprovisionInstallRunner(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationReprovisionRunner,
	})
	ctx.JSON(http.StatusOK, true)
}
