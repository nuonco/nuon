package client

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

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
		if _, err := demand.LockJob(ctx, tx, &job); err != nil {
			return err
		}
		changed, err := demand.CancelJob(ctx, tx, &job, compositeError)
		if err != nil || !changed {
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
