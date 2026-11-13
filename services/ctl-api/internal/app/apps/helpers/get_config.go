package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// GetLatestActiveAppConfig returns the app's newest active config with component config connections
// preloaded.
func (h *Helpers) GetLatestActiveAppConfig(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig

	res := h.db.WithContext(ctx).
		Scopes(
			ActiveAppConfigs(appID),
			PreloadAppConfigComponentConfigConnections,
		).
		Order("created_at DESC").
		First(&appConfig)
	if res.Error != nil {
		return nil, h.wrapAppConfigErr(ctx, res.Error)
	}

	return &appConfig, nil
}

// GetLatestActiveAppConfigBare returns the app's newest active config without preloading children.
func (h *Helpers) GetLatestActiveAppConfigBare(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig

	res := h.db.WithContext(ctx).
		Scopes(ActiveAppConfigs(appID)).
		Order("created_at DESC").
		First(&appConfig)
	if res.Error != nil {
		return nil, h.wrapAppConfigErr(ctx, res.Error)
	}

	return &appConfig, nil
}

func (h *Helpers) wrapAppConfigErr(ctx context.Context, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		desc := "This app has no app config. Run `nuon apps sync`, or run an app branch."
		if orgID, orgErr := cctx.OrgIDFromContext(ctx); orgErr == nil {
			if appSyncDisabled, flagErr := h.featuresClient.OrgHasFeature(ctx, orgID, app.OrgFeatureDisableAppSync); flagErr == nil && appSyncDisabled {
				desc = "This app has no app config. Run an app branch."
			}
		}
		return stderr.ErrNotFound{
			Err:         err,
			Description: desc,
		}
	}
	return errors.Wrap(err, "unable to get app config")
}
