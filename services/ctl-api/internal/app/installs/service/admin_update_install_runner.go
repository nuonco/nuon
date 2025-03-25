package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type AdminUpdateInstallRunnerRequest struct {
	ContainerImageTag string `json:"container_image_tag"`
}

//	@ID						AdminUpdateInstallRunner
//	@Description.markdown	update_install_runner.md
//	@Param					install_id	path	string	true	"install ID for your current install"
//	@Tags					installs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminUpdateInstallRunnerRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/installs/{install_id}/admin-update-runner [PATCH]
func (s *service) AdminUpdateInstallRunner(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminUpdateInstallRunnerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	install, err := s.adminGetInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	updates := app.RunnerGroupSettings{
		ContainerImageTag: req.ContainerImageTag,
	}
	obj := app.RunnerGroupSettings{
		RunnerGroupID: install.RunnerGroup.ID,
	}

	res := s.db.WithContext(ctx).
		Where(obj).
		Updates(updates)

	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update runner group settings: %w", res.Error))
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationReprovisionRunner,
	})

	ctx.JSON(http.StatusOK, true)
}
