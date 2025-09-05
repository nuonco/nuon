package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallWorkflowStepRequest struct {
	InstallWorkflowStepID string `json:"install_workflow_step_id" validate:"required"`
}

// @temporal-gen activity
// @by-id InstallWorkflowStepID
func (a *Activities) GetInstallWorkflowsStep(ctx context.Context, req GetInstallWorkflowStepRequest) (*app.WorkflowStep, error) {
	var step app.WorkflowStep

	res := a.db.WithContext(ctx).
		First(&step, "id = ?", req.InstallWorkflowStepID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	return &step, nil
}
