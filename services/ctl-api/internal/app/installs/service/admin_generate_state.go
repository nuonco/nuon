package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	generatestatev2 "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/generatestate"
)

// @ID						AdminInstallGenerateInstallState
// @Summary				generate state for an install via the legacy flow
// @Description.markdown	admin_install_generate_state.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/admin-generate-state [POST]
func (s *service) AdminInstallGenerateInstallState(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &generatestate.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

// @ID						AdminInstallGenerateInstallStateV2
// @Summary				generate state for an install via the v2 state manager
// @Description.markdown	admin_install_generate_state_v2.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/admin-generate-state-v2 [POST]
func (s *service) AdminInstallGenerateInstallStateV2(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallStateManagerQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get state-manager queue: %w", err))
		return
	}

	if err := s.enqueueInstallSignal(ctx, queueID, &generatestatev2.Signal{
		InstallID: install.ID,
	}, install.ID, "installs"); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue force-regenerate: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
