package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetRunner
// @Summary				get a runner by id
// @Description.markdown	get_runner.md
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners/runner,runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Runner
// @Router					/v1/runners/{runner_id} [get]
func (s *service) GetRunner(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runner)
}

func (s *service) getRunner(ctx context.Context, runnerID string) (*app.Runner, error) {
	runner := app.Runner{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}

func (s *service) getOrgRunner(ctx context.Context, runnerID string, orgID string) (*app.Runner, error) {
	runner := app.Runner{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ? AND org_id = ?", runnerID, orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
