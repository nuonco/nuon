package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						GetRunnerJobExecution
//	@Summary				get runner job execution
//	@Description.markdown	get_runner_job_execution.md
//	@Param					runner_job_id			path	string	true	"runner job ID"
//	@Param					runner_job_execution_id	path	string	true	"runner job ID"
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
//	@Success				200	{object}	app.RunnerJobExecution
//	@Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id} [get]
func (s *service) GetRunnerJobExecution(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")

	runnerJob, err := s.getRunnerJobExecution(ctx, runnerJobID, runnerJobExecutionID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job execution: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}

func (s *service) getRunnerJobExecution(ctx context.Context, runnerJobID, runnerJobExecutionID string) (*app.RunnerJobExecution, error) {
	runnerJobExecution := app.RunnerJobExecution{}
	res := s.db.WithContext(ctx).
		Preload("RunnerJob").
		First(&runnerJobExecution, "id = ?", runnerJobExecutionID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job execution: %w", res.Error)
	}

	return &runnerJobExecution, nil
}
