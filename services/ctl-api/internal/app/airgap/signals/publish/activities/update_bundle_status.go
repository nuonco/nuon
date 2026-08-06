package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateBundleStatusRequest struct {
	BundleID          string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBundleStatus(ctx context.Context, req *UpdateBundleStatusRequest) error {
	result := a.db.WithContext(ctx).Model(&app.AirgapBundle{ID: req.BundleID}).Updates(app.AirgapBundle{Status: req.Status, StatusDescription: req.StatusDescription})
	if result.Error != nil {
		return fmt.Errorf("update bundle status: %w", result.Error)
	}
	if result.RowsAffected < 1 {
		return fmt.Errorf("no bundle found: %s: %w", req.BundleID, gorm.ErrRecordNotFound)
	}
	return nil
}
