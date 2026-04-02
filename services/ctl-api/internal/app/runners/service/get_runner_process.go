package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GetRunnerProcess
// @Summary				get a runner process
// @Description.markdown	get_runner_process.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
// @Tags					runners/runner
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
func (s *service) GetRunnerProcess(ctx *gin.Context) {
	processID := ctx.Param("process_id")

	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner process: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, process)
}

func (s *service) getRunnerProcess(ctx context.Context, processID string) (*app.RunnerProcess, error) {
	var process app.RunnerProcess
	res := s.db.WithContext(ctx).
		Preload("Shutdowns").
		First(&process, "id = ?", processID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner process: %w", res.Error)
	}

	return &process, nil
}
