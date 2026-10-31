package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

type GetInstallDeployForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	ComponentID       string `validate:"required"`
	ApplyStepID       *string
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallDeployForApplyStep(ctx context.Context, req GetInstallDeployForApplyStep) (*app.InstallDeploy, error) {
	if req.ApplyStepID != nil {
		return a.getApprovedInstallDeployForApplyStep(ctx, req)
	}
	return a.getInstallDeployForApplyStep(ctx, req.InstallWorkflowID, req.ComponentID)
}

func (a *Activities) getApprovedInstallDeployForApplyStep(ctx context.Context, req GetInstallDeployForApplyStep) (*app.InstallDeploy, error) {
	ctx = routing.WithoutReplica(ctx)
	if *req.ApplyStepID == "" {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply without a workflow step", "ApplyPlanNotAuthorized", nil)
	}
	var applyStep app.WorkflowStep
	if err := a.db.WithContext(ctx).Where(app.WorkflowStep{
		ID: *req.ApplyStepID, InstallWorkflowID: req.InstallWorkflowID,
	}).First(&applyStep).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get apply step")
	}
	if applyStep.WorkflowStepGroupID == "" || applyStep.Retried {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply for an ungrouped or superseded step", "ApplyPlanNotAuthorized", nil)
	}
	var flw app.Workflow
	if err := a.db.WithContext(ctx).Where(app.Workflow{ID: applyStep.InstallWorkflowID}).First(&flw).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get apply workflow")
	}
	if flw.PlanOnly || flw.Status.Status == app.StatusCancelled {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply for a plan-only or cancelled workflow", "ApplyPlanNotAuthorized", nil)
	}
	var planStep app.WorkflowStep
	if err := a.db.WithContext(ctx).Where(&app.WorkflowStep{
		InstallWorkflowID:   applyStep.InstallWorkflowID,
		WorkflowStepGroupID: applyStep.WorkflowStepGroupID,
		GroupRetryIdx:       applyStep.GroupRetryIdx,
		ExecutionType:       app.WorkflowStepExecutionTypeApproval,
	}, "InstallWorkflowID", "WorkflowStepGroupID", "GroupRetryIdx", "ExecutionType").
		Where("idx < ?", applyStep.Idx).
		Order("idx DESC").Preload("Approval.Response").First(&planStep).Error; err != nil {
		return nil, errors.Wrap(err, "unable to find the matching plan step for apply")
	}
	if planStep.Retried || planStep.StepTargetType != string(app.WorkflowStepTargetTypeInstallDeploys) || planStep.StepTargetID == "" ||
		planStep.ResultDirective != string(directive.StepContinue) ||
		(planStep.Status.Status != app.StatusSuccess && planStep.Status.Status != app.WorkflowStepApprovalStatusApproved) {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply: the matching plan did not complete with authorization", "ApplyPlanNotAuthorized", nil)
	}
	approved := false
	if planStep.Approval != nil && planStep.Approval.Response != nil {
		approved = planStep.Approval.Response.Type == app.WorkflowStepApprovalResponseTypeApprove
	} else {
		approved = planStep.Status.Status == app.WorkflowStepApprovalStatusApproved &&
			planStep.Status.Metadata["check"] == "policy-auto-approve" &&
			planStep.Status.Metadata["check_label_approval_reason"] == "policies_passed"
	}
	if !approved {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply: the matching plan has no approval", "ApplyPlanNotAuthorized", nil)
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
