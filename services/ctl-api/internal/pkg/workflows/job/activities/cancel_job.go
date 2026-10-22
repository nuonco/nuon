package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CancelJobRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsJobCancelJob(ctx context.Context, req *CancelJobRequest) error {
	var loaded app.RunnerJob
	if err := a.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: req.ID}).First(&loaded).Error; err != nil {
		return err
	}
	if demand.SignalID(&loaded) != "" {
		_, err := a.queueClient.CancelDemandRunnerJob(ctx, req.ID, nil)
		return err
	}
	runnerJob := app.RunnerJob{
		ID: req.ID,
	}

	jobStatusV2 := app.NewCompositeStatus(ctx, app.Status(app.RunnerJobStatusCancelled))
	jobStatusV2.StatusHumanDescription = "cancelled"

	res := a.db.WithContext(ctx).
		Model(&runnerJob).
		Updates(map[string]any{
			"status":    app.RunnerJobStatusCancelled,
			"status_v2": jobStatusV2,
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
		return errors.Wrap(jres.Error, "unable to get runner job")
	}

	for _, execution := range job.Executions {
		if !execution.Status.IsRunning() {
			continue
		}

		execStatusV2 := app.NewCompositeStatus(ctx, app.Status(app.RunnerJobExecutionStatusCancelled))
		execStatusV2.StatusHumanDescription = "cancelled"

		res = a.db.WithContext(ctx).
			Model(execution).
			Updates(map[string]any{
				"status":    app.RunnerJobExecutionStatusCancelled,
				"status_v2": execStatusV2,
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to cancel job execution")
		}
		helpers.AuditJobExecutionResult(ctx, a.db, a.mw, execution.ID, app.RunnerJobExecutionStatusCancelled, "cancel_workflow")

	}

	return nil
}
