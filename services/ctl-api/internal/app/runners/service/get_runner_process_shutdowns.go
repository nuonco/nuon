package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GetRunnerProcessShutdowns
// @Summary					get shutdowns for a runner process
// @Description.markdown	get_runner_process_shutdowns.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
// @Tags					runners/runner
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.RunnerProcessShutdown
// @Router					/v1/runners/{runner_id}/processes/{process_id}/shutdowns [get]
func (s *service) GetRunnerProcessShutdowns(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")

	var shutdowns []app.RunnerProcessShutdown
	res := s.db.WithContext(ctx).
		Where("runner_process_id = ?", processID).
		Joins("JOIN runner_processes ON runner_processes.id = runner_process_shutdowns.runner_process_id AND runner_processes.runner_id = ?", runnerID).
		Find(&shutdowns)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runner process shutdowns: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, shutdowns)
}
