package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type GetInstallSandboxRunStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallSandboxRunState(ctx context.Context, req GetInstallSandboxRunStateRequest) (*app.InstallSandboxRun, error) {
	var installSandboxRun app.InstallSandboxRun
	res := a.db.WithContext(ctx).
		Scopes(
			scopes.WithOverrideTable(views.CustomViewName(a.db, &app.InstallSandboxRun{}, "state_view_v1")),
		).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").
		Preload("RunnerJobs", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_jobs_view_v2.created_at DESC")
		}).
		Preload("LogStream").
		Where(app.InstallSandboxRun{
			InstallID: req.InstallID,
			Status:    "active",
		}).
		Order("created_at desc").
		First(&installSandboxRun)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to update install action workflow run")
	}

	return &installSandboxRun, nil
}
