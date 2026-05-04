package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/generatestate"
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

	stateGenV2, err := s.useStateGenV2(ctx, install)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check state-gen-v2 feature: %w", err))
		return
	}

	if stateGenV2 {
		queueID, err := s.getInstallStateManagerQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get state-manager queue: %w", err))
			return
		}

		if err := s.enqueueInstallSignal(ctx, queueID, &stateregenerate.Signal{
			InstallID:        install.ID,
			Targets:          state.AllPartialTargets(),
			ForceAll:         true,
			TriggeredByID:    installID,
			TriggeredByType:  "installs",
			StateGeneratedBy: app.InstallStateGenerateSourceStateManager,
		}, install.ID, "installs"); err != nil {
			ctx.Error(fmt.Errorf("unable to enqueue force-regenerate: %w", err))
			return
		}
	} else {
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
			}, "", ""); err != nil {
				ctx.Error(fmt.Errorf("enqueue signal: %w", err))
				return
			}
		} else {
			s.evClient.Send(ctx, install.ID, &signals.Signal{
				Type: signals.OperationGenerateState,
			})
		}
	}

	ctx.JSON(http.StatusOK, true)
}
