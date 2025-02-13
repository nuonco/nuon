package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	WorkflowID        string        `validate:"required"`
	Status            app.ActionWorkflowStatus `validate:"required"`
	StatusDescription string        `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	aw := app.ActionWorkflow{
		ID: req.WorkflowID,
	}
	res := a.db.WithContext(ctx).Model(&aw).Updates(app.ActionWorkflow{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update ActionWorkflow: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no ActionWorkflow found: %s: %w", req.WorkflowID, gorm.ErrRecordNotFound)
	}
	return nil
}
