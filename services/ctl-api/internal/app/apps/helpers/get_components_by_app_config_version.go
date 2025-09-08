package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// This returns all app components for a point in time
func (h *Helpers) GetAppComponentsAtConfigVersion(ctx context.Context, appID string, appCfgVersion int, componentIDs []string) ([]app.Component, error) {
	var components []app.Component

	res := h.db.WithContext(ctx).
		Where(&app.Component{
			AppID: appID,
		}).
		Where("id in ?", componentIDs).
		Preload("Dependencies").
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Where("version <= ?", appCfgVersion)
		}).
		Scopes(PreloadComponentConfigConnections).
		Find(&components)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	return components, nil
}
