package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) UpdateDeployWithWorkflowID(ctx context.Context, deployID, workflowID string) error {
	var wf app.Workflow
	if err := h.db.WithContext(ctx).Select("id", "plan_only").Where(app.Workflow{ID: workflowID}).First(&wf).Error; err != nil {
		return fmt.Errorf("unable to get workflow: %w", err)
	}
	res := h.db.WithContext(ctx).Model(&app.InstallDeploy{}).
		Where("id = ?", deployID).
		Updates(map[string]any{"install_workflow_id": workflowID, "plan_only": wf.PlanOnly})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	return nil
}
