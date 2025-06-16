package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallDeployForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	ComponentID       string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetInstallDeployForApplyStep(ctx context.Context, req GetInstallDeployForApplyStep) (*app.InstallDeploy, error) {
	return a.getInstallDeployForApplyStep(ctx, req.InstallWorkflowID, req.ComponentID)
}

func (a *Activities) getInstallDeployForApplyStep(ctx context.Context, installWorkflowID, componentID string) (*app.InstallDeploy, error) {
	installDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Where(app.InstallDeploy{
			InstallWorkflowID: generics.ToPtr(installWorkflowID),
			ComponentID:       componentID,
		}).
		Preload("LogStream").
		First(&installDeploy)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install deploy")
	}

	return &installDeploy, nil
}
