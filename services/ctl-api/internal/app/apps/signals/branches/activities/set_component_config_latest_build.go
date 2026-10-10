package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"gorm.io/gorm"
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
	return buildselection.WithMutableConfig(ctx, a.db, input.AppConfigID, func(tx *gorm.DB, cfg *app.AppConfig) error {
		var build app.ComponentBuild
		if err := tx.Preload("ComponentConfigConnection").Where(app.ComponentBuild{ID: input.BuildID}).First(&build).Error; err != nil {
			return err
		}
		if build.Status != app.ComponentBuildStatusActive || build.ComponentConfigConnection.ComponentID != input.ComponentID {
			return fmt.Errorf("build %s is not a successful build for component %s", input.BuildID, input.ComponentID)
		}
		res := tx.Model(&app.ComponentConfigConnection{}).
			Where(app.ComponentConfigConnection{AppConfigID: input.AppConfigID, ComponentID: input.ComponentID}).
			Update("latest_build_id", input.BuildID)
		if res.Error != nil {
			return fmt.Errorf("unable to set latest_build_id: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("no config connection found for app_config_id=%s component_id=%s", input.AppConfigID, input.ComponentID)
		}
		return nil
	})
}
