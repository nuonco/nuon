package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): this should basically never be used at this point.
func (h *Helpers) GetAppComponentsAndLatestConfigConnection(ctx context.Context, appID string) ([]app.Component, error) {
	var components []app.Component

	res := h.db.WithContext(ctx).
		Where(&app.Component{
			AppID: appID,
		}).
		Preload("ComponentConfigs").
		Scopes(PreloadComponentConfigConnections).
		Find(&components)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	return components, nil
}
