package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminDeprovisionInstallRunnerRequest struct{}

//	@ID						AdminDeprovisionInstallRunner
//	@Description.markdown	deprovision_install_runner.md
//	@Param					install_id	path	string	true	"install ID for your current install"
//	@Tags					installs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	DeprovisionInstallRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/installs/{install_id}/admin-deprovision-runner [POST]
func (s *service) AdminDeprovisionInstallRunner(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationDeprovisionRunner,
	})
	ctx.JSON(http.StatusOK, true)
}
