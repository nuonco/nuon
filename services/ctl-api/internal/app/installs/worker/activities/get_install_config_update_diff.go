package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type GetInstallAppConfigVersionDiffInput struct {
	InstallAppConfigVersionID string `json:"install_config_update_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetInstallAppConfigVersionDiff(ctx context.Context, input *GetInstallAppConfigVersionDiffInput) (*app.InstallConfigDiff, error) {
	var update app.InstallAppConfigVersion
	if err := a.db.WithContext(ctx).First(&update, "id = ?", input.InstallAppConfigVersionID).Error; err != nil {
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

	if diff.SandboxBuildChanged && !diff.SandboxChanged {
		var builds []*app.AppSandboxBuild
		if err := a.db.WithContext(ctx).Unscoped().
			Where("id IN ?", []string{diff.SandboxBuildOldID, diff.SandboxBuildNewID}).Find(&builds).Error; err != nil {
			return nil, fmt.Errorf("unable to get sandbox builds for config diff: %w", err)
		}
		byID := make(map[string]*app.AppSandboxBuild, len(builds))
		for _, build := range builds {
			byID[build.ID] = build
		}
		diff.SandboxChanged = !configdiff.SandboxSourcesEqual(byID[diff.SandboxBuildOldID], byID[diff.SandboxBuildNewID])
	}

	return &diff, nil
}
