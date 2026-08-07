package activities

import (
	"context"
	"fmt"
)

type ReconcileInstallComponentsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallComponents(ctx context.Context, input *ReconcileInstallComponentsInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.ReconcileInstallComponents(ctx, input.InstallID)
}
