package activities

import (
	"context"
	"fmt"
)

type MigrateInstallRolesInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) MigrateInstallRoles(ctx context.Context, input *MigrateInstallRolesInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.MigrateInstallRoles(ctx, input.InstallID)
}
