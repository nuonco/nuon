package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type RestartRunnerRequest struct{}

//	@ID						AdminRestartRunner
//	@Summary				restart a runner event loop
//	@Description.markdown	restart_runner.md
//	@Param					runner_id	path	string					true	"runner ID"
//	@Param					req			body	RestartRunnerRequest	true	"Input"
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/runners/{runner_id}/restart [POST]
func (s *service) RestartRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req RestartRunnerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}
