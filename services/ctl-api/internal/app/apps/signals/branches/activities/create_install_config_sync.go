package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallConfigSyncInput struct {
	AppBranchID       string `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`
	AppBranchRunID    string `json:"app_branch_run_id,omitempty"`
	CommitSHA         string `json:"commit_sha,omitempty"`
	TriggeredBy       string `json:"triggered_by"`
	TotalInstalls     int    `json:"total_installs"`
}

type CreateInstallConfigSyncOutput struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigSync(ctx context.Context, input *CreateInstallConfigSyncInput) (*CreateInstallConfigSyncOutput, error) {
	record := app.InstallConfigSync{
		AppBranchID:       input.AppBranchID,
		AppBranchConfigID: input.AppBranchConfigID,
		CommitSHA:         input.CommitSHA,
		TriggeredBy:       input.TriggeredBy,
		TotalInstalls:     input.TotalInstalls,
		Status:            app.NewCompositeStatus(ctx, app.StatusInProgress),
	}

	if input.AppBranchRunID != "" {
		record.AppBranchRunID = &input.AppBranchRunID
	}

	if err := a.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config sync: %w", err)
	}

	return &CreateInstallConfigSyncOutput{ID: record.ID}, nil
}
