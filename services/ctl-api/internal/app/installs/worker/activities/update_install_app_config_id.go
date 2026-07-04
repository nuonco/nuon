package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallAppConfigIDInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallAppConfigID(ctx context.Context, input *UpdateInstallAppConfigIDInput) error {
	res := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("id = ?", input.InstallID).
		Update("app_config_id", input.NewAppConfigID)
	if res.Error != nil {
		return fmt.Errorf("unable to update install app_config_id: %w", res.Error)
	}
	return nil
}
