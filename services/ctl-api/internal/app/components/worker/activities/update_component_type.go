package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateComponentTypeRequest struct {
	ComponentID string            `validate:"required"`
	Type        app.ComponentType `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateComponentType(ctx context.Context, req UpdateComponentTypeRequest) error {
	cmp := app.Component{
		ID: req.ComponentID,
	}
	res := a.db.WithContext(ctx).Model(&cmp).Updates(app.Component{
		Type: req.Type,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update component: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no component found: %s %w", req.ComponentID, gorm.ErrRecordNotFound)
	}

	return nil
}
