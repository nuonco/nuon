package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// GetWorkflowsPlanOnlyMap returns a map of workflow IDs to their PlanOnly status
func (h *Helpers) GetWorkflowsPlanOnlyMap(ctx context.Context, workflowIDs []string) (map[string]bool, error) {
	if len(workflowIDs) == 0 {
		return map[string]bool{}, nil
	}

	var workflows []app.Workflow
	res := h.db.WithContext(ctx).
		Select("id, plan_only").
		Where("id IN ?", workflowIDs).
		Find(&workflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get workflows: %w", res.Error)
	}

	planOnlyMap := make(map[string]bool, len(workflows))
	for _, workflow := range workflows {
		planOnlyMap[workflow.ID] = workflow.PlanOnly
	}

	return planOnlyMap, nil
}
