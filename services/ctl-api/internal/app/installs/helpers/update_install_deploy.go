package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) UpdateDeployWithWorkflowID(ctx context.Context, deployID, workflowID string) error {
	res := h.db.WithContext(ctx).Model(&app.InstallDeploy{}).
		Where("id = ?", deployID).
		Update("install_workflow_id", workflowID)
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	return nil
}
