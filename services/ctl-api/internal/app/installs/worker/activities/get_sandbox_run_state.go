package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallSandboxRunStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetInstallSandboxRunState(ctx context.Context, req GetInstallSandboxRunStateRequest) (*app.InstallSandboxRun, error) {
	var installSandboxRun app.InstallSandboxRun
	res := a.db.WithContext(ctx).
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
		}).
		Order("updated_at desc").
		First(&installSandboxRun)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install sandbox run state")
	}

	return &installSandboxRun, nil
}
