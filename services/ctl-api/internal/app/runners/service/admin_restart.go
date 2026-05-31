package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartRunnerRequest struct{}

// @ID						AdminRestartRunner
// @Summary				restart a runner event loop
// @Description.markdown	restart_runner.md
// @Param					runner_id	path	string					true	"runner ID"
// @Param					req			body	RestartRunnerRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/runners/{runner_id}/restart [POST]
func (s *service) RestartRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req RestartRunnerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	if err := s.helpers.EnqueueRunnerSignal(ctx, runner.ID, &restart.Signal{RunnerID: runner.ID}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue restart signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
