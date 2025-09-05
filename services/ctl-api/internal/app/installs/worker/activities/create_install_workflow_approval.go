package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallWorkflowApprovalRequest struct {
	InstallWorkflowStepID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateInstallWorkflowApproval(ctx context.Context, req *CreateInstallWorkflowApprovalRequest) (*app.WorkflowStepApproval, error) {
	var step app.WorkflowStep
	res := a.db.WithContext(ctx).
		First(&step, "id = ?", req.InstallWorkflowStepID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	workflowApproval := app.WorkflowStepApproval{
		CreatedByID:           step.CreatedByID,
		InstallWorkflowStepID: step.ID,
		OrgID:                 step.OrgID,
		Type:                  app.NoopApprovalType,
	}

	resp := a.db.WithContext(ctx).Create(&workflowApproval)
	if resp.Error != nil {
		return nil, errors.Wrap(resp.Error, "unable to create workflow approval")
	}

	return &workflowApproval, nil
}
