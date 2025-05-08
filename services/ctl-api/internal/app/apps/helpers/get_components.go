package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetAppComponentsAndLatestConfigConnection(ctx context.Context, appID string) ([]app.Component, error) {
	var components []app.Component

	res := h.db.WithContext(ctx).
		Where(&app.Component{
			AppID: appID,
		}).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Scopes(PreloadComponentConfigConnections).
		Find(&components)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	return components, nil
}
