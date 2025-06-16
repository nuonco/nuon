package activities

import (
	"context"
	"errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeleteActionWorkflowRequest struct {
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ActionWorkflowID
func (a *Activities) DeleteActionWorkflow(ctx context.Context, req DeleteActionWorkflowRequest) error {
	hasMore := true

	for hasMore {
		more, err := a.BatchDeleteActionWorkflow(ctx, BatchDeleteActionWorkflowRequest{
			Limit:            20,
			ActionWorkflowID: req.ActionWorkflowID,
		})
		if err != nil {
			return temporal.NewNonRetryableApplicationError(
				"error batching delete action configs",
				"error batching delete action configs",
				err)
		}

		hasMore = more
	}

	installActionWorkflows := []app.InstallActionWorkflow{}
	res := a.db.WithContext(ctx).Delete(&installActionWorkflows, " action_workflow_id = ?", req.ActionWorkflowID)
	if res.Error != nil {
		return temporal.NewNonRetryableApplicationError(
			"error deleting install action workflows",
			"error deleting install action workflows",
			res.Error)
	}

	res = a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.ActionWorkflow{
			ID: req.ActionWorkflowID,
		})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if res.Error != nil {
		return temporal.NewNonRetryableApplicationError(
			"error deleting action workflow",
			"error deleting action workflow",
			res.Error)
	}

	return nil
}
