package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateRunnerJobExecutionRequest struct{}

//	@ID						CreateRunnerJobExecution
//	@Summary				create runner job execution
//	@Description.markdown	create_runner_job_execution.md
//	@Param					req				body	CreateRunnerJobExecutionRequest	true	"Input"
//	@Param					runner_job_id	path	string							true	"runner job ID"
//	@Tags					runners/runner
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	app.RunnerJobExecution
//	@Router					/v1/runner-jobs/{runner_job_id}/executions [POST]
func (s *service) CreateRunnerJobExecution(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner job"))
		return
	}
	cctx.SetOrgIDGinContext(ctx, runnerJob.OrgID)

	if runnerJob.ExecutionCount >= runnerJob.MaxExecutions {
		if _, err := s.cancelRunnerJob(ctx, runnerJobID); err != nil {
			ctx.Error(errors.Wrap(err, "unable to cancel runner job"))
			return
		}

		ctx.Error(fmt.Errorf("runner job has exceeded max executions"))
		return
	}

	execution, err := s.createRunnerJobExecution(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner job execution: %w", err))
		return
	}

	if err := s.updateRunnerJobStatus(ctx, runnerJobID, app.RunnerJobStatusInProgress, "in-progress"); err != nil {
		ctx.Error(errors.Wrap(err, "unable to update runner job status to in progress"))
		return
	}

	ctx.JSON(http.StatusCreated, execution)
}

func (s *service) createRunnerJobExecution(ctx context.Context, runnerJobID string) (*app.RunnerJobExecution, error) {
	runnerJobExecution := app.RunnerJobExecution{
		RunnerJobID: runnerJobID,
		Status:      app.RunnerJobExecutionStatusPending,
	}

	res := s.db.WithContext(ctx).
		Create(&runnerJobExecution)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create runner job execution: %w", res.Error)
	}

	return &runnerJobExecution, nil
}

func (s *service) updateRunnerJobStatus(ctx context.Context, runnerJobID string, runnerJobStatus app.RunnerJobStatus, runnerJobStatusDescription string) error {
	runnerJob := app.RunnerJob{
		ID: runnerJobID,
	}

	res := s.db.WithContext(ctx).
		Model(&runnerJob).
		Updates(app.RunnerJob{
			Status:            runnerJobStatus,
			StatusDescription: runnerJobStatusDescription,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to cancel runner job")
	}

	return nil
}
