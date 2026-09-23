package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateStatusRequest struct {
	ComponentID       string              `validate:"required"`
	Status            app.ComponentStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	cmp := app.Component{
		ID: req.ComponentID,
	}
	res := a.db.WithContext(ctx).Model(&cmp).Updates(app.Component{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update component")
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(
			gorm.ErrRecordNotFound,
			fmt.Sprintf("no component found: %s", req.ComponentID),
		)
	}

	return nil
}
