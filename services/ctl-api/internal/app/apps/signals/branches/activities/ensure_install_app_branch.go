package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type EnsureInstallAppBranchInput struct {
	InstallID   string `json:"install_id" validate:"required"`
	AppBranchID string `json:"app_branch_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) EnsureInstallAppBranch(ctx context.Context, input *EnsureInstallAppBranchInput) error {
	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: input.InstallID}).First(&install).Error; err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Where(app.AppBranch{ID: input.AppBranchID, AppID: install.AppID}).First(&branch).Error; err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if install.AppBranchID.Valid && install.AppBranchID.String == branch.ID {
		return nil
	}
	if err := a.helpers.SetInstallAppBranch(ctx, install.ID, branch.ID); err != nil {
		return fmt.Errorf("unable to move install to app branch: %w", err)
	}
	return nil
}
