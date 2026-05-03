package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/stateregenerate"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// @ID						AdminInstallGenerateInstallState
// @Summary				generate state for an install via the state manager
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

	queueID, err := s.getInstallStateManagerQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get state-manager queue: %w", err))
		return
	}

	triggeredByID := ctx.GetHeader("X-Nuon-Admin-Email")
	if triggeredByID == "" {
		triggeredByID = install.ID
	}

	if err := s.enqueueInstallSignal(ctx, queueID, &stateregenerate.Signal{
		InstallID:       install.ID,
		Targets:         state.AllPartialTargets(),
		ForceAll:        true,
		TriggeredByID:   triggeredByID,
		TriggeredByType: app.InstallStateGenerateSourceStateManager,
	}, install.ID, "installs"); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue force-regenerate: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
