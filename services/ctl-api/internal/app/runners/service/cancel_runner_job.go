package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type CancelRunnerJobRequest struct{}

// @ID						CancelRunnerJob
// @Summary				cancel runner job
// @Description.markdown	cancel_runner_job.md
// @Param					req				body	CancelRunnerJobRequest	true	"Input"
// @Param					runner_job_id	path	string					true	"runner job ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.RunnerJob
// @Router					/v1/runner-jobs/{runner_job_id}/cancel [POST]
func (s *service) CancelRunnerJob(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerJobID := ctx.Param("runner_job_id")

	// Verify job belongs to org before cancelling
	_, err = s.getOrgRunnerJob(ctx, runnerJobID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to cancel runner job: %w", err))
		return
	}

	runnerJob, err := s.cancelRunnerJob(ctx, runnerJobID, joberrors.CancellationReasonAPI)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to cancel runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, runnerJob)
}

func (s *service) cancelRunnerJob(ctx context.Context, runnerJobID string, reason joberrors.CancellationReason) (*app.RunnerJob, error) {
	compositeError, err := compositeerrors.New(
		&joberrors.CancellationError{Reason: reason},
		compositeerrors.WithSource("runner_jobs", runnerJobID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build runner job cancellation composite error: %w", err)
	}

	runnerJob := app.RunnerJob{
		ID:             runnerJobID,
		Status:         app.RunnerJobStatusCancelled,
		CompositeError: compositeError,
	}

	res := s.db.WithContext(ctx).
		Model(&runnerJob).
		Select("status", "composite_error").
		Updates(&runnerJob)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to cancel runner job: %w", res.Error)
	}

	job, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", err)
	}

	auditSource := "cancel_api"
	if reason == joberrors.CancellationReasonAttemptsExhausted {
		auditSource = "runner_api"
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
		helpers.AuditJobExecutionResult(ctx, s.db, s.mw, execution.ID, app.RunnerJobExecutionStatusCancelled, auditSource)

	}

	return &runnerJob, nil
}
