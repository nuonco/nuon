package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type GracefulShutdownRequest struct{}

// @ID						GracefulShutDownRunner
// @Summary				shut down a runner
// @Description.markdown	graceful_shutdown_runner.md
// @Param					runner_id	path	string							true	"runner ID"
// @Param					req			body	GracefulShutdownRequest	true	"Input"
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
// @Router					/v1/runners/{runner_id}/graceful-shutdown [POST]
func (s *service) GracefulShutDown(ctx *gin.Context) {
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

	var req GracefulShutdownRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	s.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type: signals.OperationGracefulShutdown,
	})

	ctx.JSON(http.StatusCreated, true)
}
