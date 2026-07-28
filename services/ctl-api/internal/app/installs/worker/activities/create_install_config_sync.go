package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallConfigSyncInput struct {
	InstallID              string `json:"install_id" validate:"required"`
	AppInstallConfigSyncID string `json:"app_install_config_sync_id,omitempty"`
	VCSConnectionCommitID  string `json:"vcs_connection_commit_id,omitempty"`
	AppBranchID            string `json:"app_branch_id,omitempty"`
	AppBranchConfigID      string `json:"app_branch_config_id,omitempty"`
	AppBranchRunID         string `json:"app_branch_run_id,omitempty"`
	TriggeredBy            string `json:"triggered_by"`
}

type CreateInstallConfigSyncOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigSync(ctx context.Context, input *CreateInstallConfigSyncInput) (*CreateInstallConfigSyncOutput, error) {
	record := app.InstallConfigSync{
		InstallID:         input.InstallID,
		AppBranchID:       input.AppBranchID,
		AppBranchConfigID: input.AppBranchConfigID,
		TriggeredBy:       input.TriggeredBy,
		Status:            app.NewCompositeStatus(ctx, app.StatusInProgress),
	}

	if input.AppInstallConfigSyncID != "" {
		record.AppInstallConfigSyncID = &input.AppInstallConfigSyncID
	}

	if input.VCSConnectionCommitID != "" {
		record.VCSConnectionCommitID = &input.VCSConnectionCommitID
	}

	if input.AppBranchRunID != "" {
		record.AppBranchRunID = &input.AppBranchRunID
	}

	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config sync: %w", err)
	}

	return &CreateInstallConfigSyncOutput{ID: record.ID}, nil
}
