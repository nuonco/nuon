package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// getInstallSandboxRuns reads an install's sandbox runs from the DB.
func (h *Helpers) getInstallSandboxRuns(ctx context.Context, installID string) ([]app.InstallSandboxRun, error) {
	var installSandboxRuns []app.InstallSandboxRun
	res := h.db.WithContext(ctx).
		Scopes(
			scopes.WithOverrideTable(views.CustomViewName(h.db, &app.InstallSandboxRun{}, "state_view_v1")),
		).
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
