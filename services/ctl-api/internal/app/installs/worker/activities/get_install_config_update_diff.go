package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetInstallConfigUpdateDiffInput struct {
	InstallConfigUpdateID string `json:"install_config_update_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetInstallConfigUpdateDiff(ctx context.Context, input *GetInstallConfigUpdateDiffInput) (*app.InstallConfigDiff, error) {
	var update app.InstallConfigUpdate
	if err := a.db.WithContext(ctx).First(&update, "id = ?", input.InstallConfigUpdateID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install config update: %w", err)
	}

	if update.Diff == nil {
		return &app.InstallConfigDiff{
			Added:     []app.ComponentDiffEntry{},
			Removed:   []app.ComponentDiffEntry{},
			Changed:   []app.ComponentDiffEntry{},
			Unchanged: []app.ComponentDiffEntry{},
		}, nil
	}

	diffJSON, err := update.Diff.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load diff blob: %w", err)
	}

	var diff app.InstallConfigDiff
	if err := json.Unmarshal([]byte(diffJSON), &diff); err != nil {
		return nil, fmt.Errorf("unable to parse diff: %w", err)
	}

	return &diff, nil
}
