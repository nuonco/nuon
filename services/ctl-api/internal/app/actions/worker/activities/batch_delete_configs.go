package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm/clause"
)

type BatchDeleteActionWorkflowRequest struct {
	Limit            int    `validate:"required,"`
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ActionWorkflowID
func (a *Activities) BatchDeleteActionWorkflow(ctx context.Context, req BatchDeleteActionWorkflowRequest) (bool, error) {
	hasMore := false
	configs := []app.ActionWorkflowConfig{}

	// use select to only load the necessary fields
	resp := a.db.WithContext(ctx).
		Select("id, action_workflow_id, created_at").
		Limit(req.Limit+1).
		Order("created_at ASC").
		Where("action_workflow_id = ?", req.ActionWorkflowID).
		Find(&configs)
	if resp.Error != nil {
		return hasMore, temporal.NewNonRetryableApplicationError(
			"failed to find configs",
			"failed to find configs",
			resp.Error)
	}

	hasMore = len(configs) > req.Limit

	if len(configs) > 0 {
		// configs minus one is the limit, so we can delete all but the last one
		configsToDelete := configs[:req.Limit]

		for _, config := range configsToDelete {
			installActionWorkflowRuns := []app.InstallActionWorkflowRun{}
			resp = a.db.WithContext(ctx).
				Select(clause.Associations).
				Where("action_workflow_config_id = ?", config.ID).
				Delete(&installActionWorkflowRuns)
			if resp.Error != nil {
				return hasMore, temporal.NewNonRetryableApplicationError(
					"error deleting install action workflow runs",
					"error deleting install action workflow runs",
					resp.Error)
			}

			triggers := []app.ActionWorkflowTriggerConfig{}
			resp := a.db.WithContext(ctx).
				Where("action_workflow_config_id = ?", config.ID).
				Delete(&triggers)
			if resp.Error != nil {
				return hasMore, temporal.NewNonRetryableApplicationError(
					"error deleting action workflow triggers",
					"error deleting action workflow triggers",
					resp.Error)
			}

			steps := []app.ActionWorkflowStepConfig{}
			resp = a.db.WithContext(ctx).
				Where("action_workflow_config_id = ?", config.ID).
				Delete(&steps)
			if resp.Error != nil {
				return hasMore, temporal.NewNonRetryableApplicationError(
					"error deleting action workflow steps",
					"error deleting action workflow steps",
					resp.Error)
			}

			resp = a.db.WithContext(ctx).
				Where("id = ?", config.ID).
				Delete(&config)
			if resp.Error != nil {
				return hasMore, temporal.NewNonRetryableApplicationError(
					"error deleting action workflow config",
					"error deleting action workflow config",
					resp.Error)
			}
		}
	}

	return hasMore, nil
}
