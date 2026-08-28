package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallAppConfigVersionStatusInput struct {
	ID             string            `json:"id,omitempty"`
	AppBranchRunID string            `json:"app_branch_run_id,omitempty"`
	InstallID      string            `json:"install_id" validate:"required"`
	Status         app.Status        `json:"status,omitempty"`
	StatusDesc     string            `json:"status_description,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type UpdateInstallAppConfigVersionStatusOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateInstallAppConfigVersionStatus(ctx context.Context, input *UpdateInstallAppConfigVersionStatusInput) (*UpdateInstallAppConfigVersionStatusOutput, error) {
	var version app.InstallAppConfigVersion
	query := a.db.WithContext(ctx).Where(app.InstallAppConfigVersion{InstallID: input.InstallID})
	if input.ID != "" {
		query = query.Where(app.InstallAppConfigVersion{ID: input.ID})
	} else if input.AppBranchRunID != "" {
		query = query.Where(app.InstallAppConfigVersion{AppBranchRunID: &input.AppBranchRunID})
	} else {
		return nil, fmt.Errorf("install app config version id or app branch run id is required")
	}
	if err := query.First(&version).Error; err != nil {
		return nil, fmt.Errorf("unable to find install app config version for install %s: %w", input.InstallID, err)
	}

	status := input.Status
	if status == "" {
		status = app.StatusSuccess
	}
	updates := map[string]any{
		"status": app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: input.StatusDesc,
		},
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
