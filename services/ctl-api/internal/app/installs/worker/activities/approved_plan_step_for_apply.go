package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

func (a *Activities) approvedPlanStepForApply(ctx context.Context, installWorkflowID, applyStepID string, targetType app.WorkflowStepTargetType) (*app.WorkflowStep, error) {
	if applyStepID == "" {
		return nil, temporal.NewNonRetryableApplicationError("refusing apply without a workflow step", "ApplyPlanNotAuthorized", nil)
	}
	var applyStep app.WorkflowStep
	if err := a.db.WithContext(ctx).Where(app.WorkflowStep{
		ID: applyStepID, InstallWorkflowID: installWorkflowID,
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
	if planStep.Retried || planStep.StepTargetType != string(targetType) || planStep.StepTargetID == "" ||
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
	return &planStep, nil
}
