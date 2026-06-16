package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CheckBuildNeededInput struct {
	ComponentID    string `json:"component_id"`
	NewAppConfigID string `json:"new_app_config_id"`
	OldAppConfigID string `json:"old_app_config_id"`
}

type CheckBuildNeededOutput struct {
	NeedsBuild      bool   `json:"needs_build"`
	ExistingBuildID string `json:"existing_build_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CheckBuildNeeded(ctx context.Context, input *CheckBuildNeededInput) (*CheckBuildNeededOutput, error) {
	if input.OldAppConfigID == "" {
		// No previous config to compare against — always build
		return &CheckBuildNeededOutput{NeedsBuild: true}, nil
	}

	// Load old and new component config connections for this component
	var oldConn app.ComponentConfigConnection
	err := a.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{
			AppConfigID: input.OldAppConfigID,
			ComponentID: input.ComponentID,
		}).
		First(&oldConn).Error
	if err != nil {
		// No old config for this component — it's new, needs build
		return &CheckBuildNeededOutput{NeedsBuild: true}, nil
	}

	var newConn app.ComponentConfigConnection
	err = a.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{
			AppConfigID: input.NewAppConfigID,
			ComponentID: input.ComponentID,
		}).
		First(&newConn).Error
	if err != nil {
		return &CheckBuildNeededOutput{NeedsBuild: true}, nil
	}

	// Compare checksums — if identical, the component config hasn't changed
	if oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum {
		// Find the latest successful build for the old config
		var existingBuild app.ComponentBuild
		err = a.db.WithContext(ctx).
			Where(app.ComponentBuild{
				ComponentConfigConnectionID: oldConn.ID,
				Status:                      app.ComponentBuildStatusActive,
			}).
			Order("created_at DESC").
			First(&existingBuild).Error
		if err == nil {
			return &CheckBuildNeededOutput{
				NeedsBuild:      false,
				ExistingBuildID: existingBuild.ID,
			}, nil
		}
	}

	return &CheckBuildNeededOutput{NeedsBuild: true}, nil
}
