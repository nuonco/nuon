package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
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
	}

	ctx.JSON(http.StatusOK, jobExecution)
}

func (s *service) updateRunnerJobExecution(ctx context.Context, runnerJobID, runnerJobExecutionID string, req UpdateRunnerJobExecutionRequest) (*app.RunnerJobExecution, error) {
	var job app.RunnerJob
	if err := s.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: runnerJobID}).First(&job).Error; err != nil {
		return nil, err
	}
	if demand.SignalID(&job) != "" {
		var execution app.RunnerJobExecution
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if _, err := demand.LockWorkflow(ctx, tx, job.FlowWorkflowID()); err != nil {
				return stderr.ErrConflict{Err: err, Description: "workflow is cancelled"}
			}
			if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: runnerJobID}).First(&job).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJobExecution{ID: runnerJobExecutionID, RunnerJobID: runnerJobID}).First(&execution).Error; err != nil {
				return err
			}
			if !execution.Status.IsRunning() {
				if execution.Status == req.Status {
					return nil
				}
				return stderr.ErrConflict{Err: fmt.Errorf("execution is already terminal"), Description: "execution is already terminal"}
			}
			if job.Status.IsTerminal() || !time.Now().Before(execution.CreatedAt.Add(job.ExecutionTimeout)) {
				return stderr.ErrConflict{Err: fmt.Errorf("job is terminal or expired"), Description: "job is terminal or expired"}
			}
			status := app.NewCompositeStatus(ctx, app.Status(req.Status))
			status.StatusHumanDescription = truncateStatusDescription(req.StatusDescription)
			return tx.Model(&execution).Updates(map[string]any{"status": req.Status, "status_v2": status}).Error
		})
		return &execution, err
	}
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
