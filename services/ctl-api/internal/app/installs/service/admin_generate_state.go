package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	generatestate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
)

type AdminInstallGenerateInstallStateRequest struct{}

// @ID						AdminInstallGenerateInstallState
// @Summary				generate state for an install
// @Description.markdown	admin_install_generate_state.md
// @Param					install_id	path	string						true	"install ID" // @Param					req			body	AdminInstallGenerateInstallStateRequest	false	"Input"
// @Param	req			body	AdminInstallGenerateInstallStateRequest true	"Input"
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

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, queueID, &generatestate.Signal{
			InstallID: install.ID,
		}); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationGenerateState,
		})
	}
	ctx.JSON(http.StatusOK, true)
}
