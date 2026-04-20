package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ForgetInstallRequest struct{}

// @ID						ForgetInstall
// @Summary				forget an install
// @Description.markdown	forget_install.md
// @Param					install_id	path	string					true	"install ID"
// @Param					req			body	ForgetInstallRequest	true	"Input"
// @Tags					installs
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/installs/{install_id}/forget [POST]
func (s *service) ForgetInstall(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	err = s.forgetInstall(ctx, installID)
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
		if err := s.enqueueInstallSignal(ctx, queueID, &forgotten.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationForget,
		})
	}
	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
