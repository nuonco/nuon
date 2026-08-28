package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateBundleStatusRequest struct {
	PackageID         string                   `validate:"required"`
	Status            app.ReleasePackageStatus `validate:"required"`
	StatusDescription string                   `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateBundleStatus(ctx context.Context, req *UpdateBundleStatusRequest) error {
	result := a.db.WithContext(ctx).Model(&app.ReleasePackage{ID: req.PackageID}).Updates(app.ReleasePackage{Status: req.Status, StatusDescription: req.StatusDescription})
	if result.Error != nil {
		return fmt.Errorf("update package status: %w", result.Error)
	}
	if result.RowsAffected < 1 {
		return fmt.Errorf("no package found: %s: %w", req.PackageID, gorm.ErrRecordNotFound)
	}
	return nil
}
