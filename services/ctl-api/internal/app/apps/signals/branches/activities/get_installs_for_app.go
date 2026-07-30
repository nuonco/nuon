package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field appID
func (a *Activities) getInstallsForApp(ctx context.Context, appID string) ([]app.Install, error) {
	var installs []app.Install
	if err := a.db.WithContext(ctx).
		Where(app.Install{AppID: appID}).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to get installs for app: %w", err)
	}

	return installs, nil
}
