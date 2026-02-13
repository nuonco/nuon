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

	// For install runners on VMs, send a management update signal first to pull
	// the latest image. When the management process shuts down and systemd restarts
	// the runner service, it will come back with the new image.
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
		s.evClient.Send(ctx, runner.ID, &signals.Signal{
			Type: signals.OperationMngUpdate,
		})
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationMngShutDown,
	})

	ctx.JSON(http.StatusCreated, true)
}
