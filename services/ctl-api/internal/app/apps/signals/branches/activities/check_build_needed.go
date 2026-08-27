package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

type CheckBuildNeededInput struct {
	ComponentID    string `json:"component_id"`
	NewAppConfigID string `json:"new_app_config_id"`
	OldAppConfigID string `json:"old_app_config_id"`
	SourceChanged  *bool  `json:"source_changed,omitempty"`
}

type CheckBuildNeededOutput struct {
	NeedsBuild      bool   `json:"needs_build"`
	ExistingBuildID string `json:"existing_build_id,omitempty"`
	ChangeReason    string `json:"change_reason,omitempty"`
}

const (
	ChangeReasonNoChanges     = "no_changes"
	ChangeReasonConfigChanged = "config_changed"
	ChangeReasonSourceChanged = "source_changed"
)

func buildChangeReason(needsBuild bool, oldChecksum, newChecksum string) string {
	if !needsBuild {
		return ChangeReasonNoChanges
	}
	if oldChecksum != "" && newChecksum != "" && oldChecksum != newChecksum {
		return ChangeReasonConfigChanged
	}
	return ChangeReasonSourceChanged
}

func reuseExistingBuild(existingBuildID string) *CheckBuildNeededOutput {
	return &CheckBuildNeededOutput{
		NeedsBuild:      false,
		ExistingBuildID: existingBuildID,
		ChangeReason:    ChangeReasonNoChanges,
	}
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CheckBuildNeeded(ctx context.Context, input *CheckBuildNeededInput) (*CheckBuildNeededOutput, error) {
	if input.OldAppConfigID == "" {
		return &CheckBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	var oldConn app.ComponentConfigConnection
	err := a.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{
			AppConfigID: input.OldAppConfigID,
			ComponentID: input.ComponentID,
		}).
		First(&oldConn).Error
	if err != nil {
		return &CheckBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	var newConfigConnection app.ComponentConfigConnection
	err = a.db.WithContext(ctx).
		Preload("ExternalImageComponentConfig").
		Where(app.ComponentConfigConnection{
			AppConfigID: input.NewAppConfigID,
			ComponentID: input.ComponentID,
		}).
		First(&newConfigConnection).Error
	if err != nil {
		return &CheckBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	if build.RequiresFreshBuild(&newConfigConnection) {
		return &CheckBuildNeededOutput{
			NeedsBuild:   true,
			ChangeReason: buildChangeReason(true, oldConn.Checksum, newConfigConnection.Checksum),
		}, nil
	}

	if newConfigConnection.LatestBuildID.Valid {
		var pinned app.ComponentBuild
		err = a.db.WithContext(ctx).
			Select("id", "status").
			Where(app.ComponentBuild{ID: newConfigConnection.LatestBuildID.String}).
			First(&pinned).Error
		if err == nil && pinned.Status == app.ComponentBuildStatusActive {
			return reuseExistingBuild(pinned.ID), nil
		}
		return &CheckBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	if oldConn.Checksum != "" && newConfigConnection.Checksum != "" && oldConn.Checksum == newConfigConnection.Checksum {
		if input.SourceChanged != nil && *input.SourceChanged {
			return &CheckBuildNeededOutput{
				NeedsBuild:   true,
				ChangeReason: ChangeReasonSourceChanged,
			}, nil
		}

		var existingBuild app.ComponentBuild
		err = a.db.WithContext(ctx).
			Where(app.ComponentBuild{
				ComponentConfigConnectionID: oldConn.ID,
				Status:                      app.ComponentBuildStatusActive,
			}).
			Order("created_at DESC").
			First(&existingBuild).Error
		if err == nil {
			return reuseExistingBuild(existingBuild.ID), nil
		}
	}

	return &CheckBuildNeededOutput{
		NeedsBuild:   true,
		ChangeReason: buildChangeReason(true, oldConn.Checksum, newConfigConnection.Checksum),
	}, nil
}
