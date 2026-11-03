package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

type GetInstallDeployForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	ComponentID       string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallDeployForApplyStep(ctx context.Context, req GetInstallDeployForApplyStep) (*app.InstallDeploy, error) {
	return a.getInstallDeployForApplyStep(ctx, req.InstallWorkflowID, req.ComponentID)
}

type GetApprovedInstallDeployForApplyStepRequest struct {
	InstallWorkflowID string `validate:"required"`
	ComponentID       string `validate:"required"`
	ApplyStepID       string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetApprovedInstallDeployForApplyStep(ctx context.Context, req GetApprovedInstallDeployForApplyStepRequest) (*app.InstallDeploy, error) {
	ctx = routing.WithoutReplica(ctx)
	planStep, err := a.approvedPlanStepForApply(ctx, req.InstallWorkflowID, req.ApplyStepID, app.WorkflowStepTargetTypeInstallDeploys)
	if err != nil {
		return nil, err
	}
	var deploy app.InstallDeploy
	if err := a.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where(&app.InstallDeploy{ID: planStep.StepTargetID, InstallWorkflowID: &req.InstallWorkflowID}).
		Where("install_components.component_id = ?", req.ComponentID).
		Preload("LogStream").First(&deploy).Error; err != nil {
		return nil, errors.Wrap(err, "unable to find the deploy authorized by the plan")
	}
	return &deploy, nil
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
