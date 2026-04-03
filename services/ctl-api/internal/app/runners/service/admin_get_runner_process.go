package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						AdminGetRunnerProcess
// @Summary				admin get a runner process
// @Description.markdown	admin_get_runner_process.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
// @Tags					runners/admin
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes/{process_id} [get]
func (s *service) AdminGetRunnerProcess(ctx *gin.Context) {
	processID := ctx.Param("process_id")

	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner process: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, process)
}
