package activities

import (
	"context"
	"fmt"
)

type UpdateInstallDeployWithWorkflowRequest struct {
	InstallDeployID string `validate:"required"`
	WorkflowID      string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateInstallDeployWithWorkflow(ctx context.Context, req UpdateInstallDeployWithWorkflowRequest) error {
	err := a.helpers.UpdateDeployWithWorkflowID(ctx, req.InstallDeployID, req.WorkflowID)
	if err != nil {
		return fmt.Errorf("unable to create workflow: %w", err)
	}

	return nil
}
