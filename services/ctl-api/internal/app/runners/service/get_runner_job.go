package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetRunnerJob
// @Summary				get runner job
// @Description.markdown	get_runner_job.md
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerJob
// @Router					/v1/runner-jobs/{runner_job_id} [get]
func (s *service) GetRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}

func (s *service) getRunnerJob(ctx context.Context, runnerJobID string) (*app.RunnerJob, error) {
	runnerJob := app.RunnerJob{}
	res := s.db.WithContext(ctx).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC").Limit(1)
		}).
		First(&runnerJob, "id = ?", runnerJobID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", res.Error)
	}

	return &runnerJob, nil
}

// @ID						GetRunnerJobV2
// @Summary				get runner job
// @Description.markdown	get_runner_job.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					job_id	path	string	true	"job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerJob
// @Router					/v1/runners/{runner_id}/jobs/{job_id} [get]
func (s *service) GetRunnerJobV2(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner: %w", err))
		return
	}
	runnerJobID := ctx.Param("job_id")

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}
