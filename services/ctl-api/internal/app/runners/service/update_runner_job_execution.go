package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateRunnerJobExecutionRequest struct {
	Status app.RunnerJobExecutionStatus `json:"status"`
}

//	@ID						UpdateRunnerJobExecution
//	@Summary				update a runner job execution
//	@Description.markdown	update_runner_job_execution.md
//	@Param					req						body	UpdateRunnerJobExecutionRequest	true	"Input"
//	@Param					runner_job_id			path	string							true	"runner job ID"
//	@Param					runner_job_execution_id	path	string							true	"runner job execution ID"
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
//	@Success				200	{object}	app.RunnerJobExecution
//	@Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id} [PATCH]
func (s *service) UpdateRunnerJobExecution(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")

	var req UpdateRunnerJobExecutionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	jobExecution, err := s.updateRunnerJobExecution(ctx, runnerJobID, runnerJobExecutionID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, jobExecution)
}

func (s *service) updateRunnerJobExecution(ctx context.Context, runnerJobID, runnerJobExecutionID string, req UpdateRunnerJobExecutionRequest) (*app.RunnerJobExecution, error) {
	jobExecution := app.RunnerJobExecution{}
	res := s.db.WithContext(ctx).
		Model(&jobExecution).
		Where(&app.RunnerJobExecution{
			RunnerJobID: runnerJobID,
			ID:          runnerJobExecutionID,
		}).
		Updates(app.RunnerJobExecution{
			Status: req.Status,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner job execution status: %w", res.Error)
	}

	return &jobExecution, nil
}
