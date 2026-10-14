package client

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (c *Client) CancelDemandRunnerJob(ctx context.Context, jobID string, compositeError *compositeerrors.CompositeErrorData) (*app.RunnerJob, error) {
	var job app.RunnerJob
	var resumeSignalID string
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		var flw app.Workflow
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Workflow{ID: job.FlowWorkflowID()}).First(&flw).Error; err != nil {
			return err
		}
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
			return err
		}
		if job.Status.IsTerminal() {
			return nil
		}
		job.Status, job.FinishedAt, job.CompositeError = app.RunnerJobStatusCancelled, time.Now(), compositeError
		job.StatusV2 = app.NewCompositeStatus(ctx, app.StatusCancelled)
		if err := tx.Scopes(scopes.WithDisableViews).Model(&job).Updates(map[string]any{"status": job.Status, "status_v2": job.StatusV2, "finished_at": job.FinishedAt, "composite_error": compositeError}).Error; err != nil {
			return err
		}
		if err := tx.Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: jobID}).Where("status IN ?", []app.RunnerJobExecutionStatus{app.RunnerJobExecutionStatusPending, app.RunnerJobExecutionStatusInitializing, app.RunnerJobExecutionStatusInProgress, app.RunnerJobExecutionStatusCleaningUp}).Updates(map[string]any{"status": app.RunnerJobExecutionStatusCancelled, "status_v2": job.StatusV2}).Error; err != nil {
			return err
		}
		if err := demand.UpdateOwner(ctx, tx, &job, app.StatusCancelled); err != nil {
			return err
		}
		var runner app.Runner
		if err := tx.Preload("Queues").Where(app.Runner{ID: job.RunnerID}).First(&runner).Error; err != nil {
			return err
		}
		q := runner.GetQueueForGroup(job.Group)
		if q == nil {
			return gorm.ErrRecordNotFound
		}
		raw, err := json.Marshal(map[string]any{"type": "process_job", "data": map[string]any{"runner_id": job.RunnerID, "job_id": job.ID, "demand": true, "resume": true}})
		if err != nil {
			return err
		}
		var resume signaldb.SignalData
		if err := json.Unmarshal(raw, &resume); err != nil {
			return err
		}
		resp, err := c.EnqueueSignalInTransaction(ctx, tx, &EnqueueSignalRequest{QueueID: q.ID, Signal: resume.Signal, OwnerID: job.ID, OwnerType: "runner_jobs", IdempotencyKey: "runner-demand-cancelled:" + job.ID})
		if err == nil {
			resumeSignalID = resp.ID
		}
		return err
	})
	if err == nil {
		c.DispatchSignal(resumeSignalID)
	}
	return &job, err
}
