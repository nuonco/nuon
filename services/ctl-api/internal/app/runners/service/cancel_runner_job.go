package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CancelRunnerJobRequest struct{}

//	@ID						CancelRunnerJob
//	@Summary				cancel runner job
//	@Description.markdown	cancel_runner_job.md
//	@Param					req				body	CancelRunnerJobRequest	true	"Input"
//	@Param					runner_job_id	path	string					true	"runner job ID"
//	@Tags					runners
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				202	{object}	app.RunnerJob
//	@Router					/v1/runner-jobs/{runner_job_id}/cancel [POST]
func (s *service) CancelRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.cancelRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, runnerJob)
}

func (s *service) cancelRunnerJob(ctx context.Context, runnerJobID string) (*app.RunnerJob, error) {
	runnerJob := app.RunnerJob{
		ID: runnerJobID,
	}

	res := s.db.WithContext(ctx).
		Model(&runnerJob).
		Updates(app.RunnerJob{
			Status: app.RunnerJobStatusCancelled,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to cancel runner job: %w", res.Error)
	}

	job, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", err)
	}

	for _, execution := range job.Executions {
		if !execution.Status.IsRunning() {
			continue
		}

		res = s.db.WithContext(ctx).
			Model(execution).
			Updates(app.RunnerJobExecution{
				Status: app.RunnerJobExecutionStatusCancelled,
			})
		if res.Error != nil {
			return nil, fmt.Errorf("unable to cancel job execution: %w", res.Error)
		}

	}

	return &runnerJob, nil
}
