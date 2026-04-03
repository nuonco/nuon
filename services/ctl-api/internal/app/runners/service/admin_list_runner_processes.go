package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						AdminListRunnerProcesses
// @Summary				admin list runner processes
// @Description.markdown	admin_list_runner_processes.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					type		query	string	false	"filter by process type"
// @Param					status		query	string	false	"filter by status"
// @Param					limit		query	int		false	"limit"
// @Param					offset		query	int		false	"offset"
// @Tags					runners/admin
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes [get]
func (s *service) AdminListRunnerProcesses(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	processType := ctx.Query("type")
	status := ctx.Query("status")
	limit := ctx.DefaultQuery("limit", "25")
	offset := ctx.DefaultQuery("offset", "0")

	processes, err := s.adminListRunnerProcesses(ctx, runnerID, processType, status, limit, offset)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list runner processes: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, processes)
}

func (s *service) adminListRunnerProcesses(ctx context.Context, runnerID, processType, status, limit, offset string) ([]app.RunnerProcess, error) {
	var processes []app.RunnerProcess

	query := s.db.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Preload("Shutdowns").
		Order("created_at DESC")

	if processType != "" {
		query = query.Where("type = ?", processType)
	}
	if status != "" {
		query = query.Where("composite_status->>'status' = ?", status)
	}

	query = query.Limit(intFromString(limit, 25)).Offset(intFromString(offset, 0))

	if res := query.Find(&processes); res.Error != nil {
		return nil, fmt.Errorf("unable to list runner processes: %w", res.Error)
	}

	return processes, nil
}
