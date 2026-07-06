package activities

import (
	"context"
	"fmt"
)

type MigrateInstallInputsInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	OldAppConfigID string `json:"old_app_config_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) MigrateInstallInputs(ctx context.Context, input *MigrateInstallInputsInput) error {
	installConfigMap := map[string]string{
		input.InstallID: input.OldAppConfigID,
	}

	if err := a.helpers.MigrateInstallInputsToNewConfig(ctx, a.db, installConfigMap, input.NewAppConfigID); err != nil {
		return fmt.Errorf("unable to migrate install inputs: %w", err)
	}

	return nil
}
