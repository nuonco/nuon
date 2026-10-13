package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const installGroupPlanSignalType signal.SignalType = "app-branch-plan-install-group"

type GetInstallGroupPlanResponseTypeInput struct {
	InstallWorkflowID string `json:"install_workflow_id" validate:"required"`
	InstallGroupID    string `json:"install_group_id" validate:"required"`
}

type GetInstallGroupPlanResponseTypeOutput struct {
	ResponseType app.WorkflowStepResponseType `json:"response_type"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetInstallGroupPlanResponseType(ctx context.Context, input *GetInstallGroupPlanResponseTypeInput) (*GetInstallGroupPlanResponseTypeOutput, error) {
	var steps []app.WorkflowStep
	if err := a.db.WithContext(ctx).
		Where(app.WorkflowStep{
			InstallWorkflowID: input.InstallWorkflowID,
			ExecutionType:     app.WorkflowStepExecutionTypeApproval,
		}).
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_workflow_step_approvals.created_at DESC")
		}).
		Preload("Approval.Response").
		Order("created_at DESC").
		Find(&steps).Error; err != nil {
		return nil, errors.Wrap(err, "unable to query install group plan steps")
	}

	for _, step := range steps {
		if !isInstallGroupStep(step, installGroupPlanSignalType, input.InstallGroupID) {
			continue
		}
		if step.Approval == nil || step.Approval.Response == nil {
			continue
		}
		return &GetInstallGroupPlanResponseTypeOutput{ResponseType: step.Approval.Response.Type}, nil
	}

	return &GetInstallGroupPlanResponseTypeOutput{}, nil
}
