package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunAppConfigInput struct {
	RunID       string `json:"run_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) updateAppBranchRunAppConfig(ctx context.Context, req *UpdateAppBranchRunAppConfigInput) error {
	if err := a.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where("id = ?", req.RunID).
		Update("app_config_id", req.AppConfigID); res.Error != nil {
		return fmt.Errorf("unable to update app branch run app config: %w", res.Error)
	}

	return nil
}
