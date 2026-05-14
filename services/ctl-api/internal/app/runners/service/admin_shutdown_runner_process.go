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

	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		s.l.Warn("unable to get runner process", zap.Error(err))
		ctx.Error(fmt.Errorf("unable to get runner process: %w", err))
		return
	}

	shutdown, err := s.helpers.ShutdownProcess(ctx, process, req.ShutdownType)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to shutdown runner process: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, shutdown)
}
