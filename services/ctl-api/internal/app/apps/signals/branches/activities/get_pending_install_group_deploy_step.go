package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var installGroupSkipSignalTypes = []string{
	"app-branch-update-install-group",
	"app-branch-post-deploy-runbooks",
}

type GetPendingInstallGroupDeployStepInput struct {
	InstallWorkflowID string `json:"install_workflow_id" validate:"required"`
	InstallGroupID    string `json:"install_group_id" validate:"required"`
}

type GetPendingInstallGroupDeployStepOutput struct {
	StepIDs []string `json:"step_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetPendingInstallGroupDeployStep(ctx context.Context, input *GetPendingInstallGroupDeployStepInput) (*GetPendingInstallGroupDeployStepOutput, error) {
	var steps []app.WorkflowStep
	err := a.db.WithContext(ctx).
		Where(app.WorkflowStep{
			InstallWorkflowID: input.InstallWorkflowID,
			ExecutionType:     app.WorkflowStepExecutionTypeSystem,
		}).
		Where("queue_signal->>'type' IN ?", installGroupSkipSignalTypes).
		Where("queue_signal->'data'->>'install_group_id' = ?", input.InstallGroupID).
		Where("status->>'status' IN ?", []string{
			string(app.StatusPending),
			string(app.StatusNotAttempted),
			string(app.StatusQueued),
		}).
		Order("group_idx asc, created_at asc").
		Find(&steps).Error
	if err != nil {
		return nil, errors.Wrap(err, "unable to query install group deploy step")
	}

	stepIDs := make([]string, 0, len(steps))
	for _, step := range steps {
		stepIDs = append(stepIDs, step.ID)
	}
	return &GetPendingInstallGroupDeployStepOutput{StepIDs: stepIDs}, nil
}
