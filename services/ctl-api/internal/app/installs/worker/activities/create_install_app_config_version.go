package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallAppConfigVersionInput struct {
	InstallID      string                 `json:"install_id" validate:"required"`
	OldAppConfigID string                 `json:"old_app_config_id"`
	NewAppConfigID string                 `json:"new_app_config_id" validate:"required"`
	Diff           *app.InstallConfigDiff `json:"diff,omitempty"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
	AppBranchRunID string                 `json:"app_branch_run_id,omitempty"`
	InstallGroupID string                 `json:"install_group_id,omitempty"`
}

type CreateInstallAppConfigVersionOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CreateInstallAppConfigVersion(ctx context.Context, input *CreateInstallAppConfigVersionInput) (*CreateInstallAppConfigVersionOutput, error) {
	version := app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: input.OldAppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Metadata:       input.Metadata,
		Status:         app.NewCompositeStatus(ctx, app.StatusSuccess),
	}
	if input.AppBranchRunID != "" {
		version.AppBranchRunID = &input.AppBranchRunID
	}
	if input.InstallGroupID != "" {
		version.InstallGroupID = &input.InstallGroupID
	}
	if err := a.db.WithContext(ctx).Create(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to create install app config version: %w", err)
	}

	return &CreateInstallAppConfigVersionOutput{ID: version.ID}, nil
}
