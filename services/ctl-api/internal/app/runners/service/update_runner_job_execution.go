package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type UpdateRunnerJobExecutionRequest struct {
	Status            app.RunnerJobExecutionStatus `json:"status"`
	StatusDescription string                       `json:"status_description,omitempty"`
}

// @ID						UpdateRunnerJobExecution
// @Summary				update a runner job execution
// @Description.markdown	update_runner_job_execution.md
// @Param					req						body	UpdateRunnerJobExecutionRequest	true	"Input"
// @Param					runner_job_id			path	string							true	"runner job ID"
// @Param					runner_job_execution_id	path	string							true	"runner job execution ID"
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
// @Success				200	{object}	app.RunnerJobExecution
// @Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id} [PATCH]
func (s *service) UpdateRunnerJobExecution(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")

	var req UpdateRunnerJobExecutionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	jobExecution, err := s.updateRunnerJobExecution(ctx, runnerJobID, runnerJobExecutionID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner job execution status: %w", err))
		return
	}
	helpers.AuditJobExecutionResult(ctx, s.db, s.mw, runnerJobExecutionID, req.Status, "runner_api")

	// On a terminal status, wake the process_job workflow so it finalizes
	// (stamps finished_at) on this request instead of on its next poll tick.
	// Intermediate transitions still drive the workflow via its poll loop.
	if !req.Status.IsRunning() {
		s.wakeProcessJobWorkflow(ctx, runnerJobID, processjob.TerminalSignalName(runnerJobID))
		s.closeRunnerOperationLogStream(ctx, runnerJobID)
	}

	ctx.JSON(http.StatusOK, jobExecution)
}

func (s *service) closeRunnerOperationLogStream(ctx context.Context, runnerJobID string) {
	var job app.RunnerJob
	if err := s.db.WithContext(ctx).Select("id", "log_stream_id").Where(app.RunnerJob{ID: runnerJobID}).Take(&job).Error; err != nil {
		s.l.Warn("unable to load runner job to close its log stream", zap.String("runner_job.id", runnerJobID), zap.Error(err))
		return
	}
	if job.LogStreamID == nil {
		return
	}

	res := s.db.WithContext(ctx).
		Model(&app.LogStream{}).
		Where(app.LogStream{ID: *job.LogStreamID, OwnerType: "runner_operations"}).
		Update("open", false)
	if res.Error != nil {
		s.l.Warn("unable to close runner operation log stream", zap.String("log_stream.id", *job.LogStreamID), zap.Error(res.Error))
		return
	}
	if res.RowsAffected == 0 {
		return
	}

	if err := s.acctClient.DeleteServiceAccount(ctx, *job.LogStreamID); err != nil {
		s.l.Warn("unable to delete runner operation log stream service account", zap.String("log_stream.id", *job.LogStreamID), zap.Error(err))
	}
}

func (s *service) updateRunnerJobExecution(ctx context.Context, runnerJobID, runnerJobExecutionID string, req UpdateRunnerJobExecutionRequest) (*app.RunnerJobExecution, error) {
	updates := app.RunnerJobExecution{
		Status: req.Status,
	}

	if req.StatusDescription != "" {
		current := app.RunnerJobExecution{}
		if err := s.db.WithContext(ctx).
			Where(&app.RunnerJobExecution{
				RunnerJobID: runnerJobID,
				ID:          runnerJobExecutionID,
			}).
			First(&current).Error; err != nil {
			return nil, fmt.Errorf("unable to load current job execution: %w", err)
		}

		newComposite := app.NewCompositeStatus(ctx, app.Status(req.Status))
		newComposite.StatusHumanDescription = truncateStatusDescription(req.StatusDescription)
		newComposite.History = append([]app.CompositeStatus{current.StatusV2}, current.StatusV2.History...)
		if len(newComposite.History) > 0 {
			newComposite.History[0].History = nil
		}
		updates.StatusV2 = newComposite
	}

	jobExecution := app.RunnerJobExecution{}
	res := s.db.WithContext(ctx).
		Model(&jobExecution).
		Where(&app.RunnerJobExecution{
			RunnerJobID: runnerJobID,
			ID:          runnerJobExecutionID,
		}).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner job execution status: %w", res.Error)
	}

	return &jobExecution, nil
}

// statusDescriptionMaxLen caps stored status descriptions so a long stack trace
// doesn't bloat the composite status jsonb column or Temporal history.
const statusDescriptionMaxLen = 2048

func truncateStatusDescription(s string) string {
	if len(s) <= statusDescriptionMaxLen {
		return s
	}
	return s[:statusDescriptionMaxLen] + "…(truncated)"
}
