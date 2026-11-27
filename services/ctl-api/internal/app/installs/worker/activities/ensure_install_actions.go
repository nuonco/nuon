package activities

import (
	"context"
	"fmt"
)

type EnsureInstallActionsInput struct {
	InstallID   string `json:"install_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) EnsureInstallActions(ctx context.Context, input *EnsureInstallActionsInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.EnsureInstallActions(ctx, input.InstallID, input.AppConfigID)
}
