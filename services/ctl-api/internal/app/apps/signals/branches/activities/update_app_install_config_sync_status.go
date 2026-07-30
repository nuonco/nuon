package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppInstallConfigSyncStatusInput struct {
	ID                string `json:"id" validate:"required"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateAppInstallConfigSyncStatus(ctx context.Context, input *UpdateAppInstallConfigSyncStatusInput) error {
	if err := a.db.WithContext(ctx).
		Model(&app.AppInstallConfigSync{}).
		Where("id = ?", input.ID).
		Update("status", app.CompositeStatus{
			Status:                 app.Status(input.Status),
			StatusHumanDescription: input.StatusDescription,
		}).Error; err != nil {
		return fmt.Errorf("unable to update app install config sync status: %w", err)
	}

	return nil
}
