package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID						AdminShutdownRunnerProcess
// @Summary				admin request shutdown of a runner process
// @Description.markdown	admin_shutdown_runner_process.md
// @Param					req			body	ShutdownRunnerProcessRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID"
// @Param					process_id	path	string							true	"process ID"
// @Tags					runners/admin
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerProcessShutdown
// @Router					/v1/runners/{runner_id}/processes/{process_id}/shutdown [POST]
func (s *service) AdminShutdownRunnerProcess(ctx *gin.Context) {
	processID := ctx.Param("process_id")

	var req ShutdownRunnerProcessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	shutdown, err := s.createRunnerProcessShutdown(ctx, processID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner process shutdown: %w", err))
		return
	}

	// Mark process as pending-shutdown so health checks noop
	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		s.l.Warn("unable to get runner process for pending-shutdown update", zap.Error(err))
	} else {
		if err := s.updateProcessStatusPendingShutdown(ctx, process); err != nil {
			s.l.Warn("unable to set process pending-shutdown status", zap.Error(err))
		}

		// Write a red health check to ClickHouse so dashboards reflect the shutdown
		s.createShutdownHealthCheck(ctx, process.RunnerID, processID)

		// Enqueue the process_shutdown signal to drive the shutdown lifecycle
		if err := s.helpers.EnqueueProcessShutdown(ctx, process); err != nil {
			s.l.Warn("unable to enqueue process shutdown signal", zap.Error(err))
		}
	}

	ctx.JSON(http.StatusCreated, shutdown)
}
