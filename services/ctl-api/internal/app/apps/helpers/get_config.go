package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetAppLatestConfig(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig

	res := h.db.WithContext(ctx).
		Order("created_at DESC").
		Scopes(
			PreloadAppConfigComponentConfigConnections,
		).
		First(&appConfig, "app_id = ?", appID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app config")
	}

	return &appConfig, nil
}
