package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminGracefulShutdownRequest struct{}

// @ID AdminGracefulShutDownRunner
// @Summary	shut down a runner
// @Description.markdown graceful_shutdown_runner.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminGracefulShutdownRequest	true	"Input"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/graceful-shutdown [POST]
func (s *service) AdminGracefulShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminGracefulShutdownRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type: signals.OperationGracefulShutdown,
	})

	ctx.JSON(http.StatusCreated, true)
}
