package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// getInstallSandboxRuns reads an install's sandbox runs from the DB.
func (h *Helpers) getInstallSandboxRuns(ctx context.Context, installID string) ([]app.InstallSandboxRun, error) {
	var installSandboxRuns []app.InstallSandboxRun
	res := h.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("RunnerJobs", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_jobs_view_v2.created_at DESC")
		}).
		Preload("LogStream").
		Where(app.InstallSandboxRun{
			InstallID: installID,
		}).
		Order("updated_at desc").
		Limit(5).
		Find(&installSandboxRuns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox runs: %w", res.Error)
	}

	return installSandboxRuns, nil
}
