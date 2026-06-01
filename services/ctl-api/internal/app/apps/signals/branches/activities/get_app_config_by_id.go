package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appConfigID
func (a *Activities) getAppConfigByID(ctx context.Context, appConfigID string) (*app.AppConfig, error) {
	var config app.AppConfig
	res := a.db.WithContext(ctx).
		First(&config, "id = ?", appConfigID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	return &config, nil
}
