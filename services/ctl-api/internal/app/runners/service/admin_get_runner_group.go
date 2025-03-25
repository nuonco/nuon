package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						AdminGetRunnerGroup
//	@Summary				get a runner group
//	@Description.markdown	get_runner_group.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					runner_group_id	path	string	true	"runner group ID to fetch"
//	@Produce				json
//	@Success				200	{object}	app.RunnerGroup
//	@Router					/v1/runner-groups/{runner_group_id} [GET]
func (s *service) AdminGetRunnerGroup(ctx *gin.Context) {
	runnerID := ctx.Param("runner_group_id")
	runner, err := s.getRunnerGroup(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner group: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runner)
}

func (s *service) getRunnerGroup(ctx context.Context, runnerID string) (*app.RunnerGroup, error) {
	rg := app.RunnerGroup{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Settings").
		Preload("Runners").
		First(&rg, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner group: %w", res.Error)
	}

	return &rg, nil
}
