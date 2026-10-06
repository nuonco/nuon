package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installgrouprelease"
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
		installs := input.Installs
		var existing app.InstallGroupRun
		err := a.db.WithContext(ctx).
			Select("installs").
			First(&existing, "id = ?", input.InstallGroupRunID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to get install group run: %w", err)
		}
		if err == nil {
			installs = preserveSupersededInstalls(existing.Installs, installs)
		}
		installsJSON, err := json.Marshal(installs)
		if err != nil {
			return fmt.Errorf("unable to marshal installs: %w", err)
		}
		updates["installs"] = string(installsJSON)
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

// preserveSupersededInstalls keeps a stored superseded mark when the group workflow
// writes a later completion status. The newer run records superseded on the prior
// row, and the prior group's in-memory slice never sees that write.
func preserveSupersededInstalls(stored, incoming []app.InstallGroupRunInstall) []app.InstallGroupRunInstall {
	kept := map[string]app.InstallGroupRunInstall{}
	for _, inst := range stored {
		if inst.Status == installgrouprelease.StatusSuperseded {
			kept[inst.InstallID] = inst
		}
	}
	if len(kept) == 0 {
		return incoming
	}
	out := append([]app.InstallGroupRunInstall(nil), incoming...)
	for i := range out {
		prev, ok := kept[out[i].InstallID]
		if !ok || out[i].Status == installgrouprelease.StatusSuperseded {
			continue
		}
		out[i].Status = prev.Status
		out[i].SupersededByRunID = prev.SupersededByRunID
	}
	return out
}
