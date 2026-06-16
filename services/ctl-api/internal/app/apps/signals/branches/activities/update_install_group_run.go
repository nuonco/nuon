package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallGroupRunInput struct {
	InstallGroupRunID string                       `json:"install_group_run_id" validate:"required"`
	Installs          []app.InstallGroupRunInstall `json:"installs,omitempty"`
	CompletedInstalls int                          `json:"completed_installs"`
	FailedInstalls    int                          `json:"failed_installs"`
	Status            app.CompositeStatus          `json:"status"`
	CompletedAt       *time.Time                   `json:"completed_at,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateInstallGroupRun(ctx context.Context, input *UpdateInstallGroupRunInput) error {
	updates := map[string]any{
		"status":             input.Status,
		"completed_installs": input.CompletedInstalls,
		"failed_installs":    input.FailedInstalls,
	}

	if input.Installs != nil {
		updates["installs"] = input.Installs
	}

	if input.CompletedAt != nil {
		updates["completed_at"] = input.CompletedAt
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallGroupRun{}).
		Where(app.InstallGroupRun{ID: input.InstallGroupRunID}).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update install group run: %w", res.Error)
	}

	return nil
}
