package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetCurrentRunnerProcesses
// @Summary				get current active runner processes
// @Description.markdown	get_current_runner_processes.md
// @Param					runner_id	path	string	true	"runner ID"
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
// @Success				200	{array}		app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes/current [get]
func (s *service) GetCurrentRunnerProcesses(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runnerID := ctx.Param("runner_id")

	processes, err := s.getCurrentRunnerProcesses(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get current runner processes: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, processes)
}

func (s *service) getCurrentRunnerProcesses(ctx context.Context, runnerID, orgID string) ([]app.RunnerProcess, error) {
	var processes []app.RunnerProcess

	// get the most recent active process per type using a subquery
	res := s.db.WithContext(ctx).
		Where("runner_id = ? AND org_id = ? AND composite_status->>'status' = ?", runnerID, orgID, string(app.RunnerProcessStatusActive)).
		Preload("Shutdowns").
		Order("created_at DESC").
		Find(&processes)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get current runner processes: %w", res.Error)
	}

	// deduplicate: keep only the most recent per type
	seen := make(map[app.RunnerProcessType]bool)
	var result []app.RunnerProcess
	for _, p := range processes {
		if !seen[p.Type] {
			seen[p.Type] = true
			result = append(result, p)
		}
	}

	return result, nil
}
