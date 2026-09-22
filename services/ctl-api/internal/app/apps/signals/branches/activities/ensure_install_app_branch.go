package activities

import (
	"context"
)

type EnsureInstallAppBranchInput struct {
	InstallID   string `json:"install_id" validate:"required"`
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) EnsureInstallAppBranch(ctx context.Context, input *EnsureInstallAppBranchInput) error {
	return a.installHelpers.EnsureInstallAppBranch(ctx, input.InstallID, input.AppBranchID)
}
