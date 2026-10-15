package activities

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (a *Activities) PollDemandRunnerJob(ctx context.Context, jobID, executionID string, resume signaldb.SignalData) (bool, error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("job_id", jobID), zap.String("execution_id", executionID))
	l.Debug("checking runner job execution")
	finished := false
	var resumeSignalID string
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job app.RunnerJob
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		var execution app.RunnerJobExecution
		if err := tx.Preload("Result").Where(app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).First(&execution).Error; err != nil {
			return err
		}
		now := time.Now()
		status := job.Status
		if !status.IsTerminal() {
			switch {
			case execution.Status == app.RunnerJobExecutionStatusFinished && execution.UpdatedAt.Before(execution.CreatedAt.Add(job.ExecutionTimeout)) && (execution.Result == nil || execution.Result.CreatedAt.Before(execution.CreatedAt.Add(job.ExecutionTimeout))):
				status = app.RunnerJobStatusFinished
				if execution.Result != nil && !execution.Result.Success {
					status = app.RunnerJobStatusFailed
				}
			case !execution.Status.IsRunning() && execution.Status != app.RunnerJobExecutionStatusFinished:
				status = app.RunnerJobStatusFailed
				if execution.Status == app.RunnerJobExecutionStatusCancelled {
					status = app.RunnerJobStatusCancelled
				}
				if execution.Status == app.RunnerJobExecutionStatusTimedOut {
					status = app.RunnerJobStatusTimedOut
				}
			case !now.Before(execution.CreatedAt.Add(job.ExecutionTimeout)):
				status = app.RunnerJobStatusTimedOut
				if err := tx.Model(&execution).Updates(map[string]any{"status": app.RunnerJobExecutionStatusTimedOut, "status_v2": app.NewCompositeStatus(ctx, app.Status(app.RunnerJobExecutionStatusTimedOut))}).Error; err != nil {
					return err
				}
			default:
				return nil
			}
			if err := tx.Scopes(scopes.WithDisableViews).Model(&job).Updates(map[string]any{"status": status, "status_description": string(status), "finished_at": now, "status_v2": app.NewCompositeStatus(ctx, app.Status(status))}).Error; err != nil {
				return err
			}
		}
		var runner app.Runner
		if err := tx.Preload("Queues").Where(app.Runner{ID: job.RunnerID}).First(&runner).Error; err != nil {
			return err
		}
		q := runner.GetQueueForGroup(job.Group)
		if q == nil {
			return gorm.ErrRecordNotFound
		}
		resp, err := a.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: q.ID, Signal: resume.Signal, OwnerID: jobID, OwnerType: "runner_jobs", IdempotencyKey: "runner-demand-completed:" + executionID,
		})
		if err != nil {
			return err
		}
		resumeSignalID = resp.ID
		finished = true
		return nil
	})
	if err != nil {
		l.Error("unable to check runner job execution", zap.Error(err))
	} else if finished {
		a.queueClient.DispatchSignal(resumeSignalID)
		l.Info("runner job finished and resume queued")
	}
	return finished, err
}
