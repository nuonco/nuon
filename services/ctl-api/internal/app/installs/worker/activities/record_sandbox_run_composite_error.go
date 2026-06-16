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

type RecordSandboxRunCompositeErrorRequest struct {
	SandboxRunID string `validate:"required"`
	RunnerJobID  string `validate:"required"`

	// FallbackMessage is the (possibly truncated) status description the
	// signal already has from the failed job. It is parsed when the runner
	// job execution result carries no untruncated error message.
	FallbackMessage string
}

// RecordSandboxRunCompositeError parses the failed sandbox plan/apply's
// terraform output for a recognised structured error (currently AWS IAM
// permission failures) and, when matched, freezes it onto the InstallSandboxRun
// as a CompositeError.
//
// It prefers the untruncated message stored on the runner job execution result
// (error_metadata["message"]) and falls back to the caller-supplied message.
// This activity is best-effort: a parse miss or a missing result is not an
// error, it simply leaves the run's plain-string status description in place.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordSandboxRunCompositeError(ctx context.Context, req RecordSandboxRunCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("sandbox_run_id", req.SandboxRunID),
		zap.String("runner_job_id", req.RunnerJobID),
	)

	raw := req.FallbackMessage
	if msg := a.latestJobErrorMessage(ctx, req.RunnerJobID); msg != "" {
		raw = msg
	}

	ce := deployerrors.Parse(raw)
	if ce == nil {
		l.Debug("no structured composite error recognised for sandbox run failure")
		return nil
	}

	data := compositeerrors.New(ce)
	res := a.db.WithContext(ctx).
		Model(&app.InstallSandboxRun{ID: req.SandboxRunID}).
		Select("composite_error").
		Updates(app.InstallSandboxRun{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to record sandbox run composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no sandbox run found for id %s: %w", req.SandboxRunID, gorm.ErrRecordNotFound)
	}

	l.Info("recorded composite error on sandbox run", zap.String("composite_error_type", string(data.Type)))
	return nil
}
