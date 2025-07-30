package service

import (
	"context"
	"encoding/base64"
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

	// compressed versions
	ContentsCompressed        string `json:"contents_compressed" swaggertype:"string"`
	ContentsDisplayCompressed string `json:"contents_display_compressed" swaggertype:"string"`
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

	// branch on wether or not the content received is compressed.
	var jobExecution *app.RunnerJobExecutionResult
	if req.Contents != "" {
		jobExecution, err := s.createRunnerJobExecutionResult(ctx, runnerJobID, runnerJobExecutionID, &req)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
			return
		}

		jobExecution.ContentsDisplay = nil
		jobExecution.Contents = ""
	}
	if req.ContentsCompressed != "" {
		jobExecution, err := s.createRunnerJobExecutionResultFromCompressed(ctx, runnerJobID, runnerJobExecutionID, &req)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
			return
		}

		jobExecution.ContentsDisplay = nil
		jobExecution.Contents = ""
	}

	ctx.JSON(http.StatusCreated, jobExecution)
}

func (s *service) createRunnerJobExecutionResultFromCompressed(ctx context.Context, runnerJobID, runnerJobExecutionID string, req *CreateRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, err
	}

	contentsGzip, err := base64.URLEncoding.DecodeString(req.ContentsCompressed)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode contents")
	}
	contentsDisplayGzip, err := base64.URLEncoding.DecodeString(req.ContentsDisplayCompressed)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode contents display")
	}
	result := app.RunnerJobExecutionResult{
		OrgID:                runnerJob.OrgID,
		RunnerJobExecutionID: runnerJobExecutionID,
		Success:              req.Success,
		ContentsGzip:         contentsGzip,
		ContentsDisplayGzip:  contentsDisplayGzip,
		ErrorCode:            req.ErrorCode,
		ErrorMetadata:        pgtype.Hstore(req.ErrorMetadata),
	}

	res := s.db.WithContext(ctx).Create(&result)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to write runner job execution result: %w")
	}

	return &result, nil
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
		Contents:             req.Contents,
		ErrorCode:            req.ErrorCode,
		ErrorMetadata:        pgtype.Hstore(req.ErrorMetadata),
	}

	res := s.db.WithContext(ctx).Create(&result)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to write runner job execution result: %w")
	}

	if req.ContentsDisplay != nil {
		// NOTE(fd): we split up the write because this column can be rather large.
		// TODO(fd): return a 206 partial content, ensure the client knows how to handle it.
		byts, err := json.Marshal(req.ContentsDisplay)
		if err != nil {
			return nil, errors.Wrap(res.Error, "unable to marshal contents display")
		}
		rjer := app.RunnerJobExecutionResult{
			ID: result.ID,
		}
		updateRes := s.db.WithContext(ctx).Model(&rjer).Updates(app.RunnerJobExecutionResult{
			ContentsDisplay: byts,
		})
		if updateRes.Error != nil {
			return &result, errors.Wrap(res.Error, "failed to set display content on runner job execution")
		}
	}

	return &result, nil
}
