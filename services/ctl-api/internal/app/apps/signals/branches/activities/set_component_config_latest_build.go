package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SetComponentConfigLatestBuildInput struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
	ComponentID string `json:"component_id" validate:"required"`
	BuildID     string `json:"build_id" validate:"required"`
}

// SetComponentConfigLatestBuild pins latest_build_id on the CCC for the given
// app config + component. Used when CheckBuildNeeded skips a rebuild so the
// new CCC still points at the reused Active build.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) SetComponentConfigLatestBuild(ctx context.Context, input *SetComponentConfigLatestBuildInput) error {
	res := a.db.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(app.ComponentConfigConnection{
			AppConfigID: input.AppConfigID,
			ComponentID: input.ComponentID,
		}).
		Update("latest_build_id", input.BuildID)
	if res.Error != nil {
		return fmt.Errorf("unable to set latest_build_id: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("no config connection found for app_config_id=%s component_id=%s", input.AppConfigID, input.ComponentID)
	}
	return nil
}
