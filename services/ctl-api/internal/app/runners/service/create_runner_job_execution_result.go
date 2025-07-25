package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateRunnerJobExecutionResultRequest struct {
	Success bool `json:"success"`

	ErrorMetadata map[string]*string `json:"error_metadata"`
	ErrorCode     int                `json:"error_code"`

	Contents        string                 `json:"contents" swaggertype:"string"`
	ContentsDisplay map[string]interface{} `json:"contents_display"`
}

// @ID						CreateRunnerJobExecutionResult
// @Summary				create a runner job execution result
// @Description.markdown	create_runner_job_execution_result.md
// @Param					req						body	CreateRunnerJobExecutionResultRequest	true	"Input"
// @Param					runner_job_id			path	string									true	"runner job ID"
// @Param					runner_job_execution_id	path	string									true	"runner job execution ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerJobExecutionResult
// @Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id}/result [POST]
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

	jobExecution.ContentsDisplay = nil
	jobExecution.Contents = ""
	ctx.JSON(http.StatusCreated, jobExecution)
}

func (s *service) createRunnerJobExecutionResult(ctx context.Context, runnerJobID, runnerJobExecutionID string, req *CreateRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, err
	}

	byts, err := json.Marshal(req.ContentsDisplay)
	result := app.RunnerJobExecutionResult{
		OrgID:                runnerJob.OrgID,
		RunnerJobExecutionID: runnerJobExecutionID,
		Success:              req.Success,
		Contents:             req.Contents,
		ErrorCode:            req.ErrorCode,
		ErrorMetadata:        pgtype.Hstore(req.ErrorMetadata),
	}

	res := s.db.WithContext(ctx).Create(&result)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to write runner job execution result: %w")
	}

	// NOTE(fd): we split up the write because this column can be rather large.
	// TODO(fd): return a 206 partial content, ensure the client knows how to handle it.

	rjer := app.RunnerJobExecutionResult{
		ID: result.ID,
	}
	updateRes := s.db.WithContext(ctx).Model(&rjer).Updates(app.RunnerJobExecutionResult{
		ContentsDisplay: byts,
	})
	if updateRes.Error != nil {
		return &result, errors.Wrap(res.Error, "failed to set display content on runner job execution")
	}

	return &result, nil
}
