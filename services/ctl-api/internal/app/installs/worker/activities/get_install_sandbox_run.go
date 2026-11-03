package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

type GetInstallSandboxRunForApplyStep struct {
	InstallWorkflowID string `validate:"required"`
	InstallID         string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallSandboxRunForApplyStep(ctx context.Context, req GetInstallSandboxRunForApplyStep) (*app.InstallSandboxRun, error) {
	return a.getInstallSandboxRun(ctx, req.InstallWorkflowID, req.InstallID)
}

type GetApprovedInstallSandboxRunForApplyStepRequest struct {
	InstallWorkflowID string `validate:"required"`
	InstallID         string `validate:"required"`
	ApplyStepID       string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetApprovedInstallSandboxRunForApplyStep(ctx context.Context, req GetApprovedInstallSandboxRunForApplyStepRequest) (*app.InstallSandboxRun, error) {
	ctx = routing.WithoutReplica(ctx)
	if req.ApplyStepID == "" {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply without a workflow step", "ApplyPlanNotAuthorized", nil)
	}
	var applyStep app.WorkflowStep
	if err := a.db.WithContext(ctx).Where(app.WorkflowStep{
		ID: req.ApplyStepID, InstallWorkflowID: req.InstallWorkflowID,
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
		Order("idx DESC").
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "install_workflow_step_id")
		}).
		Preload("Approval.Response").First(&planStep).Error; err != nil {
		return nil, errors.Wrap(err, "unable to find the matching plan step for apply")
	}
	if planStep.Retried || planStep.StepTargetType != string(app.WorkflowStepTargetTypeInstallSandboxRuns) || planStep.StepTargetID == "" ||
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
	var run app.InstallSandboxRun
	if err := a.db.WithContext(ctx).
		Where(&app.InstallSandboxRun{ID: planStep.StepTargetID, InstallID: req.InstallID}).
		Preload("LogStream").First(&run).Error; err != nil {
		return nil, errors.Wrap(err, "unable to find the sandbox run authorized by the plan")
	}
	return &run, nil
}

type GetLatestInstallSandboxRun struct {
	InstallWorkflowID string `validate:"required"`
	InstallID         string `validate:"required"`
}

// @temporal-gen-v2 activity
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
