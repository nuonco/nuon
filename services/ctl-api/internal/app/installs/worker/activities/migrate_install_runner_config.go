package activities

import (
	"context"
	"fmt"
)

type MigrateInstallRunnerConfigInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) MigrateInstallRunnerConfig(ctx context.Context, input *MigrateInstallRunnerConfigInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.MigrateInstallRunnerConfig(ctx, input.InstallID)
}
