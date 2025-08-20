package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallSandboxRunForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	InstallID         string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetInstallSandboxRunForApplyStep(ctx context.Context, req GetInstallSandboxRunForApplyStep) (*app.InstallSandboxRun, error) {
	return a.getInstallSandboxRun(ctx, req.InstallWorkflowID, req.InstallID)
}

type GetLatestInstallSandboxRun struct {
	InstallWorkflowID string `validate:"required"`
	InstallID         string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetLatestInstallSandboxRun(ctx context.Context, req GetInstallSandboxRunForApplyStep) (*app.InstallSandboxRun, error) {
	return a.getInstallSandboxRun(ctx, req.InstallWorkflowID, req.InstallID)
}

func (a *Activities) getInstallSandboxRun(ctx context.Context, installWorkflowID, installID string) (*app.InstallSandboxRun, error) {
	installSandboxRun := app.InstallSandboxRun{}
	res := a.db.WithContext(ctx).
		Where(app.InstallSandboxRun{
			InstallWorkflowID: generics.ToPtr(installWorkflowID),
			InstallID:         installID,
		}).
		Preload("LogStream").
		Order("created_at desc").
		First(&installSandboxRun)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install sandbox run")
	}

	return &installSandboxRun, nil
}
