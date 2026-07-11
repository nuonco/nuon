package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallAppConfigVersionStatusInput struct {
	AppBranchRunID string            `json:"app_branch_run_id" validate:"required"`
	InstallID      string            `json:"install_id" validate:"required"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type UpdateInstallAppConfigVersionStatusOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallAppConfigVersionStatus(ctx context.Context, input *UpdateInstallAppConfigVersionStatusInput) (*UpdateInstallAppConfigVersionStatusOutput, error) {
	var version app.InstallAppConfigVersion
	if err := a.db.WithContext(ctx).
		Where(app.InstallAppConfigVersion{
			InstallID: input.InstallID,
		}).
		Where("app_branch_run_id = ?", input.AppBranchRunID).
		First(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to find install app config version for run %s install %s: %w", input.AppBranchRunID, input.InstallID, err)
	}

	updates := map[string]any{
		"status": app.NewCompositeStatus(ctx, app.StatusSuccess),
	}
	if input.Metadata != nil {
		updates["metadata"] = input.Metadata
	}

	if err := a.db.WithContext(ctx).
		Model(&version).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("unable to update install app config version status: %w", err)
	}

	return &UpdateInstallAppConfigVersionStatusOutput{ID: version.ID}, nil
}
