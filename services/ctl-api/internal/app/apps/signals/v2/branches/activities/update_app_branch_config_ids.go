package activities

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchConfigIDsInput struct {
	AppBranchConfigID string   `json:"app_branch_config_id" validate:"required"`
	ComponentIDs      []string `json:"component_ids"`
	ActionIDs         []string `json:"action_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) updateAppBranchConfigIDs(ctx context.Context, req *UpdateAppBranchConfigIDsInput) error {
	if err := a.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	updates := map[string]interface{}{
		"component_ids": pq.StringArray(req.ComponentIDs),
		"action_ids":    pq.StringArray(req.ActionIDs),
	}

	if res := a.db.WithContext(ctx).Model(&app.AppBranchConfig{}).
		Where("id = ?", req.AppBranchConfigID).
		Updates(updates); res.Error != nil {
		return fmt.Errorf("unable to update app branch config IDs: %w", res.Error)
	}

	return nil
}
