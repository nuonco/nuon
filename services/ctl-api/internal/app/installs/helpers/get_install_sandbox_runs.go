package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// getInstallSandboxRuns reads an install's sandbox runs from the DB.
func (h *Helpers) getInstallSandboxRuns(ctx context.Context, installID string) ([]app.InstallSandboxRun, error) {
	var installSandboxRuns []app.InstallSandboxRun
	res := h.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("RunnerJob").
		Preload("LogStream").
		Where("install_id = ?", installID).
		Order("created_at desc").
		Limit(5).
		Find(&installSandboxRuns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox runs: %w", res.Error)
	}

	return installSandboxRuns, nil
}
