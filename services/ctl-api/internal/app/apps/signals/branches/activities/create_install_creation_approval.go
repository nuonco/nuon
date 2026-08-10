package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallCreationApprovalInput struct {
	AppID                  string                `json:"app_id" validate:"required"`
	AppInstallConfigSyncID string                `json:"app_install_config_sync_id" validate:"required"`
	ProposedInstalls       []app.ProposedInstall `json:"proposed_installs" validate:"required"`
}

type CreateInstallCreationApprovalOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallCreationApproval(ctx context.Context, input *CreateInstallCreationApprovalInput) (*CreateInstallCreationApprovalOutput, error) {
	record := app.InstallCreationApproval{
		AppID:                  input.AppID,
		AppInstallConfigSyncID: input.AppInstallConfigSyncID,
		ProposedInstalls:       input.ProposedInstalls,
		Status:                 app.InstallCreationApprovalStatusPending,
	}

	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("unable to create install creation approval: %w", err)
	}

	return &CreateInstallCreationApprovalOutput{ID: record.ID}, nil
}
