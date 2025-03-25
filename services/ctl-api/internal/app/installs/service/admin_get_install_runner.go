package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@ID						AdminGetInstallRunner
//	@Summary				get an install runner
//	@Description.markdown	get_install_runner_group.md
//	@Param					install_id	path	string	true	"install ID"
//	@Tags					installs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/installs/{install_id}/admin-get-runner [GET]
func (s *service) AdminGetInstallRunner(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.adminGetInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install.RunnerGroup.Runners[0])
}
