package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field installID
func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).
		Preload("App").
		Preload("AppConfig").
		First(&install, "id = ?", installID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	return &install, nil
}
