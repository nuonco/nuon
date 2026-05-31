package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/forceshutdown"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminForceShutdownRequest struct{}

// @ID						AdminForceShutDownRunner
// @Summary				force shut down a runner
// @Description.markdown	force_shutdown_runner.md
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	AdminForceShutdownRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/force-shutdown [POST]
func (s *service) AdminForceShutDown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminForceShutdownRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &forceshutdown.Signal{RunnerID: runnerID}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue force-shutdown signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
