package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/gracefulshutdown"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminGracefulShutdownRequest struct{}

// @ID						AdminGracefulShutDownRunner
// @Summary				shut down a runner
// @Description.markdown	graceful_shutdown_runner.md
// @Param					runner_id	path	string							true	"runner ID"
// @Param					req			body	AdminGracefulShutdownRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/graceful-shutdown [POST]
func (s *service) AdminGracefulShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminGracefulShutdownRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &gracefulshutdown.Signal{RunnerID: runnerID}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue graceful-shutdown signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
