package installcreate

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const branchConfigPageSize = 100

// ResolveInputConfig returns the input config belonging to the app config that
// install creation will pin. Apps without branches retain the legacy app-wide
// latest-input behavior.
func ResolveInputConfig(ctx context.Context, api nuon.Client, appID, appBranchID string) (*models.AppAppInputConfig, error) {
	if appBranchID == "" {
		return api.GetAppInputLatestConfig(ctx, appID)
	}

	configs, _, err := api.GetAppBranchAppConfigs(ctx, appID, appBranchID, &models.GetPaginatedQuery{
		Limit: branchConfigPageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list app configs for branch %s: %w", appBranchID, err)
	}

	var selected *models.AppAppConfig
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
	if selected == nil {
		return nil, fmt.Errorf("selected app branch %s has no active non-preview app config", appBranchID)
	}

	recurse := true
	cfg, err := api.GetAppConfig(ctx, appID, selected.ID, &recurse)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config %s for branch %s: %w", selected.ID, appBranchID, err)
	}
	if cfg.Input == nil {
		return nil, fmt.Errorf("app config %s for branch %s has no input config", selected.ID, appBranchID)
	}
	return cfg.Input, nil
}
