package activities

import (
	"context"
	"fmt"
)

type ReconcileInstallActionsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallActions(ctx context.Context, input *ReconcileInstallActionsInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.ReconcileInstallActions(ctx, input.InstallID)
}
