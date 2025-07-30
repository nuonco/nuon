package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// GetPinnedAppInputConfig gets the pinned input config for an install.
func (h *Helpers) GetPinnedAppInputConfig(ctx context.Context, appID, appConfigID string) (*app.AppInputConfig, error) {
	parentAppConfig := app.AppConfig{}
	res := h.db.WithContext(ctx).
		Where(app.AppConfig{
			AppID: appID,
		}).
		Preload("InputConfig").
		Preload("InputConfig.AppInputs").
		First(&parentAppConfig, "id = ?", appConfigID)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get app input config: %w", res.Error)
	}

	return &parentAppConfig.InputConfig, nil
}
