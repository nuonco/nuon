package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type RecordJobLifecycleCompositeErrorRequest struct {
	JobID  string                           `validate:"required"`
	Reason joberrors.LifecycleFailureReason `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordJobLifecycleCompositeError(ctx context.Context, req RecordJobLifecycleCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("job_id", req.JobID),
		zap.String("reason", string(req.Reason)),
	)

	var job app.RunnerJob
	res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Select("id", "owner_id", "owner_type").
		Where(app.RunnerJob{ID: req.JobID}).
		Take(&job)
	if res.Error != nil {
		return fmt.Errorf("unable to get runner job for lifecycle composite error: %w", res.Error)
	}

	if req.Reason == joberrors.LifecycleFailureReasonResultMissing {
		var execution app.RunnerJobExecution
		res = a.db.WithContext(ctx).
			Select("id").
			Where(app.RunnerJobExecution{RunnerJobID: req.JobID}).
			Order("created_at DESC").
			Order("id DESC").
			Take(&execution)
		if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to get latest runner job execution: %w", res.Error)
		}
		if res.Error == nil {
			var result app.RunnerJobExecutionResult
			res = a.db.WithContext(ctx).
				Select("id").
				Where(app.RunnerJobExecutionResult{RunnerJobExecutionID: execution.ID}).
				Take(&result)
			if res.Error == nil {
				l.Info("runner job execution result exists, skipping missing-result composite error")
				return nil
			}
			if res.Error != gorm.ErrRecordNotFound {
				return fmt.Errorf("unable to check for runner job execution result: %w", res.Error)
			}
		}
	}

	data, err := compositeerrors.New(
		&joberrors.LifecycleFailureError{Reason: req.Reason},
		compositeerrors.WithSource("runner_jobs", req.JobID),
	)
	if err != nil {
		return fmt.Errorf("unable to build runner job lifecycle composite error: %w", err)
	}

	res = a.db.WithContext(ctx).
		Model(&job).
		Select("composite_error").
		Updates(app.RunnerJob{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to record runner job lifecycle composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner job found for id %s: %w", req.JobID, gorm.ErrRecordNotFound)
	}

	if job.OwnerType != "install_deploys" && job.OwnerType != "install_sandbox_runs" {
		l.Info("recorded runner job lifecycle composite error")
		return nil
	}

	ownerCompositeError, err := runnershelpers.GetLatestJobCompositeError(ctx, a.db, runnershelpers.GetLatestJobCompositeErrorRequest{
		OwnerID:   job.OwnerID,
		OwnerType: job.OwnerType,
	})
	if err != nil {
		l.Warn("unable to resolve runner job lifecycle composite error for owner", zap.Error(err))
		return nil
	}

	var mirrorRes *gorm.DB
	switch job.OwnerType {
	case "install_deploys":
		mirrorRes = a.db.WithContext(ctx).
			Model(&app.InstallDeploy{ID: job.OwnerID}).
			Select("composite_error").
			Updates(app.InstallDeploy{CompositeError: ownerCompositeError})
	case "install_sandbox_runs":
		mirrorRes = a.db.WithContext(ctx).
			Model(&app.InstallSandboxRun{ID: job.OwnerID}).
			Select("composite_error").
			Updates(app.InstallSandboxRun{CompositeError: ownerCompositeError})
	}
	if mirrorRes != nil && (mirrorRes.Error != nil || mirrorRes.RowsAffected < 1) {
		l.Warn("unable to mirror runner job lifecycle composite error to owner",
			zap.String("owner_type", job.OwnerType),
			zap.String("owner_id", job.OwnerID),
			zap.Error(mirrorRes.Error))
	}

	l.Info("recorded runner job lifecycle composite error")
	return nil
}
