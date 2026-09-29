package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

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
		return nil, wrapAppConfigErr(res.Error)
	}

	return &appConfig, nil
}

func (h *Helpers) GetLatestActiveAppConfigBare(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig

	res := h.db.WithContext(ctx).
		Scopes(ActiveAppConfigs(appID)).
		Order("created_at DESC").
		First(&appConfig)
	if res.Error != nil {
		return nil, wrapAppConfigErr(res.Error)
	}

	return &appConfig, nil
}

func wrapAppConfigErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return stderr.ErrNotFound{
			Err:         err,
			Description: "App has no successfully synced config, please sync the app config first",
		}
	}
	return errors.Wrap(err, "unable to get app config")
}
