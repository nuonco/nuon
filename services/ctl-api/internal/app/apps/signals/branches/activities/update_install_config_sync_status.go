package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallConfigSyncStatusInput struct {
	InstallConfigSyncID string `json:"install_config_sync_id" validate:"required"`
	SyncedInstalls      int    `json:"synced_installs"`
	FailedInstalls      int    `json:"failed_installs"`
	Status              string `json:"status"`
	StatusDescription   string `json:"status_description"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateInstallConfigSyncStatus(ctx context.Context, input *UpdateInstallConfigSyncStatusInput) error {
	updates := map[string]any{
		"synced_installs": input.SyncedInstalls,
		"failed_installs": input.FailedInstalls,
		"status": app.CompositeStatus{
			Status:                 app.Status(input.Status),
			StatusHumanDescription: input.StatusDescription,
		},
	}

	if err := a.db.WithContext(ctx).
		Model(&app.InstallConfigSync{}).
		Where("id = ?", input.InstallConfigSyncID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("unable to update install config sync status: %w", err)
	}

	return nil
}
