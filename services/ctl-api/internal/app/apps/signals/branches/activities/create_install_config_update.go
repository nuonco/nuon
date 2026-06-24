package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallConfigUpdateInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
	AppBranchRunID string `json:"app_branch_run_id" validate:"required"`
	InstallGroupID string `json:"install_group_id"`
}

type CreateInstallConfigUpdateOutput struct {
	InstallConfigUpdateID string                 `json:"install_config_update_id"`
	Diff                  *app.InstallConfigDiff `json:"diff,omitempty"`
	InstallName           string                 `json:"install_name,omitempty"`
	InstallLabels         map[string]string      `json:"install_labels,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigUpdate(ctx context.Context, input *CreateInstallConfigUpdateInput) (*CreateInstallConfigUpdateOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	diff, err := a.computeInstallConfigDiff(ctx, install.AppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallConfigUpdate{
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if err := a.db.WithContext(ctx).Create(&update).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	if err := a.saveDiffBlob(ctx, update.ID, diff); err != nil {
		a.l.Warn("unable to save config diff blob", zap.Error(err))
	}

	return &CreateInstallConfigUpdateOutput{
		InstallConfigUpdateID: update.ID,
		Diff:                  diff,
		InstallName:           install.Name,
		InstallLabels:         install.Labels,
	}, nil
}
