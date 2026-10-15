package activities

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (a *Activities) PublishRunnerJob(ctx context.Context, signalID, jobID string) (app.RunnerJobStatus, error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("signal_id", signalID), zap.String("job_id", jobID))
	l.Info("publishing prepared runner job")
	var status app.RunnerJobStatus
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var qs app.QueueSignal
		if err := tx.Where(app.QueueSignal{ID: signalID}).First(&qs).Error; err != nil {
			return err
		}
		if qs.Status.Status == app.StatusCancelled {
			return temporal.NewNonRetryableApplicationError("signal is cancelled", "CANCELLED", nil)
		}
		var step app.WorkflowStep
		if qs.OwnerType != "install_workflow_steps" {
			return fmt.Errorf("runner demand requires a workflow step")
		}
		if err := tx.Where(app.WorkflowStep{ID: qs.OwnerID}).First(&step).Error; err != nil {
			return err
		}
		if _, err := demand.LockWorkflow(ctx, tx, step.InstallWorkflowID); err != nil {
			return temporal.NewNonRetryableApplicationError("workflow is unavailable", "CANCELLED", err)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.QueueSignal{ID: signalID}).First(&qs).Error; err != nil {
			return err
		}
		if qs.Status.Status == app.StatusCancelled {
			return temporal.NewNonRetryableApplicationError("signal is cancelled", "CANCELLED", nil)
		}
		if err := tx.Where(app.WorkflowStep{ID: step.ID}).First(&step).Error; err != nil {
			return err
		}
		if step.Status.Status == app.StatusCancelled {
			return temporal.NewNonRetryableApplicationError("step is cancelled", "CANCELLED", nil)
		}
		var job app.RunnerJob
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		status = job.Status
		if job.Status.IsTerminal() || job.Status == app.RunnerJobStatusInProgress {
			return nil
		}
		var plan app.RunnerJobPlan
		if err := tx.Where(app.RunnerJobPlan{RunnerJobID: jobID}).First(&plan).Error; err != nil {
			return fmt.Errorf("job is not prepared: %w", err)
		}
		if len(qs.Continuation) == 0 {
			return fmt.Errorf("job has no continuation checkpoint")
		}
		metadata := job.Metadata
		if metadata == nil {
			metadata = pgtype.Hstore{}
		}
		metadata[demand.SignalKey], metadata[demand.StepKey] = &signalID, &step.ID
		metadata["flow_workflow_id"] = &step.InstallWorkflowID
		status = app.RunnerJobStatusAvailable
		updates := map[string]any{
			"metadata": metadata, "status": status, "status_description": "waiting for runner", "status_v2": app.NewCompositeStatus(ctx, app.StatusWaitingForRunner),
			"queue_timeout": 0, "available_timeout": 0, "overall_timeout": 0,
		}
		var runner app.Runner
		if err := tx.Preload("RunnerGroup.Settings").Preload("Org.CreatedBy").Where(app.Runner{ID: job.RunnerID}).First(&runner).Error; err != nil {
			return err
		}
		if runner.RunnerGroup.Settings.SandboxMode && (runner.Org.CreatedBy.AccountType == app.AccountTypeCanary || !a.cfg.SandboxModeEnableRunners) {
			now := time.Now()
			exec := app.RunnerJobExecution{RunnerJobID: jobID, Status: app.RunnerJobExecutionStatusFinished, OrgID: job.OrgID, CreatedByID: job.CreatedByID}
			if err := tx.Create(&exec).Error; err != nil {
				return err
			}
			if err := tx.Create(&app.RunnerJobExecutionResult{RunnerJobExecutionID: exec.ID, Success: true, OrgID: job.OrgID, CreatedByID: job.CreatedByID}).Error; err != nil {
				return err
			}
			status = app.RunnerJobStatusFinished
			updates["status"], updates["started_at"], updates["finished_at"] = status, now, now
			updates["status_v2"] = app.NewCompositeStatus(ctx, app.StatusSuccess)
		}
		if err := tx.Scopes(scopes.WithDisableViews).Model(&job).Updates(updates).Error; err != nil {
			return err
		}
		if status == app.RunnerJobStatusFinished {
			return nil
		}
		if err := tx.Model(&step).Updates(map[string]any{
			"runner_job_id": jobID, "runner_job_signal_id": signalID,
			"status": app.NewCompositeStatus(ctx, app.StatusWaitingForRunner), "result_directive": demand.AwaitRunner, "finished_at": nil, "started_at": nil,
		}).Error; err != nil {
			return err
		}
		return demand.UpdateOwner(ctx, tx, &job, app.StatusWaitingForRunner)
	})
	if err != nil {
		l.Error("unable to publish runner job", zap.Error(err))
	} else {
		l.Info("runner job published", zap.String("status", string(status)))
	}
	return status, err
}
