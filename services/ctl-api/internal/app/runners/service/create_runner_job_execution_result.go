package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateRunnerJobExecutionResultRequest struct {
	Success bool `json:"success"`

	ErrorMetadata map[string]*string `json:"error_metadata"`
	ErrorCode     int                `json:"error_code"`
}

//	@ID						CreateRunnerJobExecutionResult
//	@Summary				create a runner job execution result
//	@Description.markdown	create_runner_job_execution_result.md
//	@Param					req						body	CreateRunnerJobExecutionResultRequest	true	"Input"
//	@Param					runner_job_id			path	string									true	"runner job ID"
//	@Param					runner_job_execution_id	path	string									true	"runner job execution ID"
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
//	@Success				201	{object}	app.RunnerJobExecutionResult
//	@Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id}/result [POST]
func (s *service) CreateRunnerJobExecutionResult(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")

	var req CreateRunnerJobExecutionResultRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	jobExecution, err := s.createRunnerJobExecutionResult(ctx, runnerJobID, runnerJobExecutionID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, jobExecution)
}

func (s *service) createRunnerJobExecutionResult(ctx context.Context, runnerJobID, runnerJobExecutionID string, req *CreateRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, err
	}

	result := app.RunnerJobExecutionResult{
		OrgID:                runnerJob.OrgID,
		RunnerJobExecutionID: runnerJobExecutionID,
		Success:              req.Success,

		ErrorCode:     req.ErrorCode,
		ErrorMetadata: pgtype.Hstore(req.ErrorMetadata),
	}

	res := s.db.WithContext(ctx).
		Create(&result)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to write runner job execution result: %w", res.Error)
	}

	return &result, nil
}
