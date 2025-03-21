package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminForceShutdownRequest struct{}

// @ID AdminForceShutDownRunner
// @Summary	force shut down a runner
// @Description.markdown force_shutdown_runner.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminForceShutdownRequest	true	"Input"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/force-shutdown [POST]
func (s *service) AdminForceShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminForceShutdownRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationForceShutdown,
	})

	ctx.JSON(http.StatusCreated, true)
}
