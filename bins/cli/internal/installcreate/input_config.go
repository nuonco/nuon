package installcreate

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const branchConfigPageSize = 100

// InputConfigForInstall returns the input config the install is pinned to.
// That is the branch config when the install is on an app branch, which can
// declare inputs the app's latest config does not. An install with no pinned
// app config falls back to the latest input config.
func InputConfigForInstall(ctx context.Context, api nuon.Client, install *models.AppInstall) (*models.AppAppInputConfig, error) {
	if install == nil {
		return nil, fmt.Errorf("install is required")
	}
	if install.AppConfigID == "" {
		return api.GetAppInputLatestConfig(ctx, install.AppID)
	}

	recurse := true
	cfg, err := api.GetAppConfig(ctx, install.AppID, install.AppConfigID, &recurse)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config %s: %w", install.AppConfigID, err)
	}
	if cfg == nil || cfg.Input == nil {
		return nil, fmt.Errorf("app config %s has no input config", install.AppConfigID)
	}
	hydrateInputGroups(cfg.Input)
	return cfg.Input, nil
}

// ResolveInputConfig returns the input config belonging to the app config that
// install creation will pin. Apps without branches retain the legacy app-wide
// latest-input behavior.
func ResolveInputConfig(ctx context.Context, api nuon.Client, appID, appBranchID string) (*models.AppAppInputConfig, error) {
	if appBranchID == "" {
		return api.GetAppInputLatestConfig(ctx, appID)
	}

	branch, err := api.GetAppBranch(ctx, appID, appBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch %s: %w", appBranchID, err)
	}
	if _, err := AwaitBranchConfig(ctx, api, appID, branch, false, ui.PrintLn); err != nil {
		return nil, err
	}

	var selected *models.AppAppConfig
	query := &models.GetPaginatedQuery{Limit: branchConfigPageSize}
	for {
		configs, hasNext, err := api.GetAppBranchAppConfigs(ctx, appID, appBranchID, query)
		if err != nil {
			return nil, fmt.Errorf("unable to list app configs for branch %s: %w", appBranchID, err)
		}
		for _, cfg := range configs {
			if cfg == nil || cfg.ID == "" {
				continue
			}
			if cfg.Status != models.AppAppConfigStatusActive {
				continue
			}
			if cfg.Labels["source"] == string(models.AppAppBranchRunTypeGitDashPreviewDashRun) {
				continue
			}
			selected = cfg
			break
		}
		if selected != nil || !hasNext || len(configs) == 0 {
			break
		}
		query.Offset += len(configs)
	}
	if selected == nil {
		return nil, fmt.Errorf("selected app branch %s has no active non-preview app config (no successful branch run to pin)", appBranchID)
	}

	recurse := true
	cfg, err := api.GetAppConfig(ctx, appID, selected.ID, &recurse)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config %s for branch %s: %w", selected.ID, appBranchID, err)
	}
	if cfg.Input == nil {
		return nil, fmt.Errorf("app config %s for branch %s has no input config", selected.ID, appBranchID)
	}
	hydrateInputGroups(cfg.Input)
	return cfg.Input, nil
}

// hydrateInputGroups mirrors the nested shape returned by the latest-input
// endpoint. Full app configs preload inputs as a flat list, while the install
// creator renders inputs from their groups.
func hydrateInputGroups(inputConfig *models.AppAppInputConfig) {
	if inputConfig == nil || len(inputConfig.InputGroups) == 0 || len(inputConfig.Inputs) == 0 {
		return
	}

	groupsByID := make(map[string]*models.AppAppInputGroup, len(inputConfig.InputGroups))
	for _, group := range inputConfig.InputGroups {
		if group == nil {
			continue
		}
		group.AppInputs = nil
		groupsByID[group.ID] = group
	}
	for _, input := range inputConfig.Inputs {
		if input == nil {
			continue
		}
		if group := groupsByID[input.GroupID]; group != nil {
			group.AppInputs = append(group.AppInputs, input)
		}
	}
}
