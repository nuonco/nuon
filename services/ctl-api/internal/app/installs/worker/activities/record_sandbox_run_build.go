package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RecordSandboxRunBuildRequest struct {
	SandboxRunID string `json:"sandbox_run_id" validate:"required"`
	BuildID      string `json:"build_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) RecordSandboxRunBuild(ctx context.Context, req RecordSandboxRunBuildRequest) error {
	result := a.db.WithContext(ctx).Model(&app.InstallSandboxRun{}).
		Where(app.InstallSandboxRun{ID: req.SandboxRunID}).
		Update("app_sandbox_build_id", req.BuildID)
	if result.Error != nil {
		return fmt.Errorf("record sandbox run build: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("record sandbox run build: sandbox run %s not found", req.SandboxRunID)
	}
	return nil
}
