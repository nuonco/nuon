package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type RecordDeployCompositeErrorRequest struct {
	DeployID    string `validate:"required"`
	RunnerJobID string `validate:"required"`

	// FallbackMessage is the (possibly truncated) status description the
	// signal already has from the failed job. It is parsed when the runner
	// job execution result carries no untruncated error message.
	FallbackMessage string
}

// RecordDeployCompositeError parses the failed deploy's terraform output for a
// recognised structured error (currently AWS IAM permission failures) and, when
// matched, freezes it onto the InstallDeploy as a CompositeError.
//
// It prefers the untruncated message stored on the runner job execution result
// (error_metadata["message"]) and falls back to the caller-supplied message.
// This activity is best-effort: a parse miss or a missing result is not an
// error, it simply leaves the deploy's plain-string status description in place.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordDeployCompositeError(ctx context.Context, req RecordDeployCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("deploy_id", req.DeployID),
		zap.String("runner_job_id", req.RunnerJobID),
	)

	raw := req.FallbackMessage
	if msg := a.latestJobErrorMessage(ctx, req.RunnerJobID); msg != "" {
		raw = msg
	}

	ce := deployerrors.Parse(raw)
	if ce == nil {
		l.Debug("no structured composite error recognised for deploy failure")
		return nil
	}

	data := compositeerrors.New(ce)
	res := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{ID: req.DeployID}).
		Select("composite_error").
		Updates(app.InstallDeploy{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to record deploy composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no deploy found for id %s: %w", req.DeployID, gorm.ErrRecordNotFound)
	}

	l.Info("recorded composite error on deploy", zap.String("composite_error_type", string(data.Type)))
	return nil
}

// latestJobErrorMessage returns the untruncated error message recorded on the
// most recent runner job execution result for the job, or "" when none exists.
// It walks the RunnerJob -> Executions -> Result association chain (the
// executions table is indexed on (runner_job_id, created_at)).
func (a *Activities) latestJobErrorMessage(ctx context.Context, runnerJobID string) string {
	var job app.RunnerJob
	res := a.db.WithContext(ctx).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC, runner_job_executions.id DESC").Limit(1)
		}).
		Preload("Executions.Result").
		Where(app.RunnerJob{ID: runnerJobID}).
		First(&job)
	if res.Error != nil {
		return ""
	}

	if len(job.Executions) == 0 || job.Executions[0].Result == nil {
		return ""
	}

	if msg, ok := job.Executions[0].Result.ErrorMetadata["message"]; ok && msg != nil {
		return *msg
	}
	return ""
}
