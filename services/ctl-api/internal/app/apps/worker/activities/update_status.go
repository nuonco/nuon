package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	AppID             string        `validate:"required"`
	Status            app.AppStatus `validate:"required"`
	StatusDescription string        `validate:"required"`
}

func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	currentApp := app.App{
		ID: req.AppID,
	}
	res := a.db.WithContext(ctx).Model(&currentApp).Updates(app.App{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update app: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no app found: %s: %w", req.AppID, gorm.ErrRecordNotFound)
	}
	return nil
}
