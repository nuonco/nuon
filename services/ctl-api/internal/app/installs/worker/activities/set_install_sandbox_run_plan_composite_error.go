package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/stackerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetInstallSandboxRunPlanCompositeErrorRequest struct {
	SandboxRunID string `validate:"required"`
	// Detail is the sanitised plan-render error message. Empty clears the field.
	Detail string
}

// SetInstallSandboxRunPlanCompositeError freezes a SandboxPlanRenderError onto
// the given sandbox run row. Passing an empty Detail clears the field (nil),
// which is used at the start of a retry to remove a stale error.
//
// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetInstallSandboxRunPlanCompositeError(ctx context.Context, req SetInstallSandboxRunPlanCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("sandbox_run_id", req.SandboxRunID))

	var data *compositeerrors.CompositeErrorData
	if req.Detail != "" {
		var err error
		data, err = compositeerrors.New(
			&stackerrors.SandboxPlanRenderError{Detail: req.Detail},
			compositeerrors.WithSource("install_sandbox_runs", req.SandboxRunID),
		)
		if err != nil {
			return fmt.Errorf("unable to build sandbox plan render composite error: %w", err)
		}
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallSandboxRun{ID: req.SandboxRunID}).
		Select("composite_error").
		Updates(app.InstallSandboxRun{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to set install sandbox run plan composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no sandbox run found for id %s: %w", req.SandboxRunID, gorm.ErrRecordNotFound)
	}

	l.Info("updated install sandbox run plan composite error")
	return nil
}
