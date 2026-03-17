package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field appConfigID
func (a *Activities) getAppBranchRunByAppConfigID(ctx context.Context, appConfigID string) (*app.AppBranchRun, error) {
	if appConfigID == "" {
		return nil, fmt.Errorf("appConfigID is required")
	}

	var run app.AppBranchRun
	res := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Where("app_config_id = ?", appConfigID).
		First(&run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app branch run for app config %s: %w", appConfigID, res.Error)
	}

	return &run, nil
}
