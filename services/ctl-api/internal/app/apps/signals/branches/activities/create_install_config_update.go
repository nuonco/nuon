package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallAppConfigVersionInput struct {
	InstallID      string `json:"install_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
	AppBranchRunID string `json:"app_branch_run_id" validate:"required"`
	InstallGroupID string `json:"install_group_id"`
}

type CreateInstallAppConfigVersionOutput struct {
	InstallAppConfigVersionID string                 `json:"install_config_update_id"`
	Diff                      *app.InstallConfigDiff `json:"diff,omitempty"`
	InstallName               string                 `json:"install_name,omitempty"`
	InstallLabels             map[string]string      `json:"install_labels,omitempty"`
	OldAppConfigID            string                 `json:"old_app_config_id,omitempty"`
	NewAppConfigID            string                 `json:"new_app_config_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallAppConfigVersion(ctx context.Context, input *CreateInstallAppConfigVersionInput) (*CreateInstallAppConfigVersionOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	diff, err := a.computeInstallConfigDiff(ctx, install.AppConfigID, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallAppConfigVersion{
		AppBranchRunID: &input.AppBranchRunID,
		InstallGroupID: &input.InstallGroupID,
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

	return &CreateInstallAppConfigVersionOutput{
		InstallAppConfigVersionID: update.ID,
		Diff:                      diff,
		InstallName:               install.Name,
		InstallLabels:             install.Labels,
		OldAppConfigID:            install.AppConfigID,
		NewAppConfigID:            input.NewAppConfigID,
	}, nil
}
