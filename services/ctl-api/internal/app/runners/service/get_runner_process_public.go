package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetRunnerProcessPublic
// @Summary				get a runner process
// @Description.markdown	get_runner_process_public.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
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
// @Success				200	{object}	app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes/{process_id} [get]
func (s *service) GetRunnerProcessPublic(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	processID := ctx.Param("process_id")

	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner process: %w", err))
		return
	}

	if process.OrgID != org.ID {
		ctx.Error(fmt.Errorf("runner process not found"))
		return
	}

	ctx.JSON(http.StatusOK, process)
}
