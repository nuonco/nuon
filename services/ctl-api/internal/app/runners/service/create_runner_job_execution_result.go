package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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

	// Source identity for image-type builds.
	// These fields are populated by the containerimage build handler after it
	// resolves the user-provided source ref to a manifest descriptor. When the
	// runner job's owner is a ComponentBuild and SourceDigest is non-empty,
	// the result handler persists these fields onto the ComponentBuild row.
	//
	// SourceRef is what the user wrote in the spec ("nginx:1.25.3" or
	// "nginx@sha256:..."). Always populated for image-type builds.
	SourceRef string `json:"source_ref,omitempty"`
	// SourceImage is the repository portion of SourceRef (no tag/digest).
	SourceImage string `json:"source_image,omitempty"`
	// ResolvedTag is the tag the runner pulled from. Empty for digest-pinned
	// refs.
	ResolvedTag string `json:"resolved_tag,omitempty"`
	// SourceDigest is the manifest list digest of the resolved source ref.
	// Used for build dedup against the prior Active build.
	SourceDigest string `json:"source_digest,omitempty"`
	// SourceMediaType is the media type of the resolved manifest.
	SourceMediaType string `json:"source_media_type,omitempty"`
	// ResolvedAt is when the runner resolved SourceRef to SourceDigest.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	// NoOp is true when the resolved SourceDigest matched the previous build's
	// SourceDigest and the runner skipped the artifact push.
	NoOp bool `json:"no_op,omitempty"`
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerJobExecutionResult
// @Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id}/result [POST]
func (s *service) CreateRunnerJobExecutionResult(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")

	var req CreateRunnerJobExecutionResultRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// branch on wether or not the content received is compressed.
	var jobExecution *app.RunnerJobExecutionResult
	var err error
	if req.ContentsCompressed != "" || req.ContentsDisplayCompressed != "" {
		jobExecution, err = s.createRunnerJobExecutionResultFromCompressed(ctx, runnerJobID, runnerJobExecutionID, &req)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
			return
		}
	} else {
		jobExecution, err = s.createRunnerJobExecutionResult(ctx, runnerJobID, runnerJobExecutionID, &req)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
			return
		}
	}

	if err := s.applyComponentBuildSourceIdentity(ctx, runnerJobID, &req); err != nil {
		// Non-fatal: source-identity persistence shouldn't fail the result
		// write. The build's status workflow runs independently.
		s.l.Warn("unable to apply component build source identity", zap.Error(err))
	}

	ctx.JSON(http.StatusCreated, jobExecution)
}

// applyComponentBuildSourceIdentity persists source-identity fields
// onto the ComponentBuild row when the runner job's owner is a ComponentBuild
// and the runner reported a SourceDigest. Skipped otherwise.
func (s *service) applyComponentBuildSourceIdentity(ctx context.Context, runnerJobID string, req *CreateRunnerJobExecutionResultRequest) error {
	if req.SourceDigest == "" {
		return nil
	}

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	// Only image-type builds emit source identity; the owner type is the
	// ComponentBuild table name. Matches the value used when the build job is
	// created in components/worker/activities/create_build_job.go.
	if runnerJob.OwnerType != "component_builds" {
		return nil
	}
	if runnerJob.OwnerID == "" {
		return nil
	}

	build := app.ComponentBuild{ID: runnerJob.OwnerID}
	updates := app.ComponentBuild{
		SourceRef:       req.SourceRef,
		SourceImage:     req.SourceImage,
		ResolvedTag:     req.ResolvedTag,
		SourceDigest:    req.SourceDigest,
		SourceMediaType: req.SourceMediaType,
		ResolvedAt:      req.ResolvedAt,
		NoOp:            req.NoOp,
	}
	res := s.db.WithContext(ctx).
		Model(&build).
		Select("source_ref", "source_image", "resolved_tag", "source_digest", "source_media_type", "resolved_at", "no_op").
		Updates(updates)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update component build source identity")
	}
	return nil
}

func (s *service) createRunnerJobExecutionResultFromCompressed(ctx context.Context, runnerJobID, runnerJobExecutionID string, req *CreateRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, err
	}

	// Runner sends gzip-compressed payloads encoded as base64 strings.
	// We decode once here and persist the raw gzip bytes for later decompression.
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
