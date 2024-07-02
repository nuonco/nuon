package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	ComponentID       string              `validate:"required"`
	Status            app.ComponentStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	cmp := app.Component{
		ID: req.ComponentID,
	}
	res := a.db.WithContext(ctx).Model(&cmp).Updates(app.Component{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update component: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no component found: %s %w", req.ComponentID, gorm.ErrRecordNotFound)
	}

	return nil
}
