package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateSandboxBuildStatusRequest struct {
	BuildID           string                    `json:"build_id" validate:"required"`
	Status            app.AppSandboxBuildStatus `json:"status" validate:"required"`
	StatusDescription string                    `json:"status_description"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateSandboxBuildStatus(ctx context.Context, req UpdateSandboxBuildStatusRequest) error {
	res := a.db.WithContext(ctx).Model(&app.AppSandboxBuild{ID: req.BuildID}).Updates(map[string]interface{}{
		"status":             req.Status,
		"status_description": req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update sandbox build status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("sandbox build not found %s: %w", req.BuildID, gorm.ErrRecordNotFound)
	}
	return nil
}
