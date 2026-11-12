package installcreate

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func createInstallInputConfig(ctx context.Context, api nuon.Client, appID string) (*models.AppAppInputConfig, error) {
	cfg, err := api.GetAppInstallInputConfig(ctx, appID, "")
	if nuon.IsMissingRoute(err) {
		return api.GetAppInputLatestConfig(ctx, appID)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get install input config for app %s: %w", appID, err)
	}
	hydrateInputGroups(cfg)
	return cfg, nil
}
