package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type MngShutDownRequest struct{}

// @ID						ShutDownRunnerMng
// @Summary				shut down an install runner management process
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	MngShutDownRequest	true	"Input"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{boolean}	true
// @Router					/v1/runners/{runner_id}/mng/shutdown [POST]
func (s *service) MngShutDown(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerID := ctx.Param("runner_id")
	runner, err := s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner %s: %w", runnerID, err))
		return
	}

	var req MngShutDownRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	// For install runners on VMs, send a graceful shutdown to the install runner
	// process instead of just restarting the management process. The install runner's
	// job loop will finish any in-flight work before picking up the shutdown job.
	// When the runner exits, systemd Restart=always kicks in and ExecStartPre does
	// docker pull, which fetches the latest image for the floating tag (e.g. main).
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
		s.evClient.Send(ctx, runner.ID, &signals.Signal{
			Type: signals.OperationGracefulShutdown,
		})
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationMngShutDown,
	})

	ctx.JSON(http.StatusCreated, true)
}
