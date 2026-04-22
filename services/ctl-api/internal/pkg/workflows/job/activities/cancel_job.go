package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/processjobsignals"
	"go.temporal.io/sdk/activity"
)

type CancelJobRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsJobCancelJob(ctx context.Context, req *CancelJobRequest) error {
	runnerJob := app.RunnerJob{
		ID: req.ID,
	}

	res := a.db.WithContext(ctx).
		Model(&runnerJob).
		Updates(app.RunnerJob{
			Status: app.RunnerJobStatusCancelled,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to cancel runner job")
	}

	job := app.RunnerJob{}
	jres := a.db.WithContext(ctx).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC").Limit(1)
		}).
		First(&job, "id = ?", req.ID)
	if jres.Error != nil {
		return errors.Wrap(res.Error, "unable to get runner job")
	}

	for _, execution := range job.Executions {
		if !execution.Status.IsRunning() {
			continue
		}

		res = a.db.WithContext(ctx).
			Model(execution).
			Updates(app.RunnerJobExecution{
				Status: app.RunnerJobExecutionStatusCancelled,
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to cancel job execution")
		}
	}

	l := activity.GetLogger(ctx)
	if err := a.tClient.SignalWorkflowInNamespace(ctx,
		signals.TemporalNamespace,
		processjobsignals.WorkflowID(req.ID),
		"",
		processjobsignals.SignalName,
		processjobsignals.WakeUp{Reason: processjobsignals.ReasonJobStatusChanged},
	); err != nil {
		l.Debug("processjob cancel signal skipped", zap.Error(err))
	}

	return nil
}
