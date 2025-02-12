package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeleteActionWorkflowRequest struct {
	WorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id WorkflowID
func (a *Activities) DeleteActionWorkflow(ctx context.Context, req DeleteActionWorkflowRequest) error {
	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.ActionWorkflow{
		ID: req.WorkflowID,
	})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if res.Error != nil {
		return fmt.Errorf("unable to delete ActionWorkflow: %w", res.Error)
	}

	return nil
}
