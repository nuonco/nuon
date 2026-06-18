package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallDeployForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	ComponentID       string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallDeployForApplyStep(ctx context.Context, req GetInstallDeployForApplyStep) (*app.InstallDeploy, error) {
	return a.getInstallDeployForApplyStep(ctx, req.InstallWorkflowID, req.ComponentID)
}

func (a *Activities) getInstallDeployForApplyStep(ctx context.Context, installWorkflowID, componentID string) (*app.InstallDeploy, error) {
	installDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_deploys.install_workflow_id = ?", installWorkflowID).
		Where("install_components.component_id = ?", componentID).
		Preload("LogStream").
		Order("install_deploys.created_at desc").
		First(&installDeploy)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install deploy")
	}

	return &installDeploy, nil
}
