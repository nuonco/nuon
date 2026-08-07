package activities

import (
	"context"
	"fmt"
)

type ReconcileInstallRunbooksInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallRunbooks(ctx context.Context, input *ReconcileInstallRunbooksInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return a.helpers.ReconcileInstallRunbooks(ctx, input.InstallID)
}
