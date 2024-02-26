package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateConfigStatusRequest struct {
	AppConfigID       string              `validate:"required"`
	Status            app.AppConfigStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

func (a *Activities) UpdateConfigStatus(ctx context.Context, req UpdateConfigStatusRequest) error {
	appCfg := app.AppConfig{
		ID: req.AppConfigID,
	}
	res := a.db.WithContext(ctx).Model(&appCfg).Updates(app.AppConfig{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update app config: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no app config found: %s: %w", req.AppConfigID, gorm.ErrRecordNotFound)
	}

	return nil
}
