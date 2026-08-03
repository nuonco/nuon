package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetPendingInstallGroupDeployStepInput struct {
	InstallWorkflowID string `json:"install_workflow_id" validate:"required"`
	InstallGroupID    string `json:"install_group_id" validate:"required"`
}

type GetPendingInstallGroupDeployStepOutput struct {
	StepID string `json:"step_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetPendingInstallGroupDeployStep(ctx context.Context, input *GetPendingInstallGroupDeployStepInput) (*GetPendingInstallGroupDeployStepOutput, error) {
	var step app.WorkflowStep
	err := a.db.WithContext(ctx).
		Where(app.WorkflowStep{
			InstallWorkflowID: input.InstallWorkflowID,
			ExecutionType:     app.WorkflowStepExecutionTypeSystem,
		}).
		Where("queue_signal->'data'->>'install_group_id' = ?", input.InstallGroupID).
		Where("status->>'status' IN ?", []string{
			string(app.StatusPending),
			string(app.StatusNotAttempted),
			string(app.StatusQueued),
		}).
		Order("group_idx desc, created_at desc").
		First(&step).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GetPendingInstallGroupDeployStepOutput{}, nil
		}
		return nil, errors.Wrap(err, "unable to query install group deploy step")
	}

	return &GetPendingInstallGroupDeployStepOutput{StepID: step.ID}, nil
}
