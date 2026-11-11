package installcreate

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const unbranchedConfigPageSize = 100

// UnbranchedInputConfig returns the input config of the app's newest active
// config created by nuon apps sync, skipping app-branch configs and
// git-preview-run branch previews. It is the CLI's equivalent of the server's
// unbranched install pin, used when an install has no app_branch.
func UnbranchedInputConfig(ctx context.Context, api nuon.Client, appID string) (*models.AppAppInputConfig, error) {
	query := &models.GetPaginatedQuery{Limit: unbranchedConfigPageSize}
	for {
		configs, hasNext, err := api.GetAppConfigs(ctx, appID, query)
		if err != nil {
			return nil, fmt.Errorf("unable to list app configs for app %s: %w", appID, err)
		}
		for _, cfg := range configs {
			if cfg == nil || cfg.ID == "" || cfg.AppBranchID != "" {
				continue
			}
			if cfg.Status != models.AppAppConfigStatusActive {
				continue
			}
			if cfg.Labels["source"] == string(models.AppAppBranchRunTypeGitDashPreviewDashRun) {
				continue
			}

			recurse := true
			full, err := api.GetAppConfig(ctx, appID, cfg.ID, &recurse)
			if err != nil {
				return nil, fmt.Errorf("unable to get app config %s: %w", cfg.ID, err)
			}
			if full == nil || full.Input == nil {
				continue
			}
			hydrateInputGroups(full.Input)
			return full.Input, nil
		}
		if !hasNext || len(configs) == 0 {
			break
		}
		query.Offset += len(configs)
	}

	return nil, noUnbranchedConfigError(ctx, api, appID)
}

func noUnbranchedConfigError(ctx context.Context, api nuon.Client, appID string) error {
	branches, _, err := api.GetAppBranches(ctx, appID, &models.GetPaginatedQuery{Limit: 1})
	if err == nil && len(branches) > 0 {
		return fmt.Errorf("no active app config from nuon apps sync; this app's configs are on app branches; set app_branch to create the install on an app branch")
	}
	return fmt.Errorf("no active app config from nuon apps sync; run nuon apps sync first")
}
