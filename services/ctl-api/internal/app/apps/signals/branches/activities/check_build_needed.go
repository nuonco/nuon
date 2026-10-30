package activities

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

type CheckBuildNeededInput struct {
	ComponentID    string `json:"component_id"`
	NewAppConfigID string `json:"new_app_config_id"`
	OldAppConfigID string `json:"old_app_config_id"`
	RunID          string `json:"run_id,omitempty"`
	SourceChanged  *bool  `json:"source_changed,omitempty"`
}

type CheckBuildNeededOutput struct {
	NeedsBuild      bool                     `json:"needs_build"`
	ExistingBuildID string                   `json:"existing_build_id,omitempty"`
	ChangeReason    string                   `json:"change_reason,omitempty"`
	SourceIdentity  *sourceidentity.Identity `json:"source_identity,omitempty"`
}

const (
	ChangeReasonNoChanges       = "no_changes"
	ChangeReasonConfigChanged   = "config_changed"
	ChangeReasonSourceChanged   = "source_changed"
	ChangeReasonSourceAndConfig = "source_and_config"
)

func configChecksumChanged(oldChecksum, newChecksum string) bool {
	return oldChecksum != "" && newChecksum != "" && oldChecksum != newChecksum
}
func sourceCommitChanged(input *CheckBuildNeededInput) bool {
	return input != nil && input.SourceChanged != nil && *input.SourceChanged
}
func buildChangeReason(sourceChanged bool, oldChecksum, newChecksum string) string {
	switch {
	case sourceChanged && configChecksumChanged(oldChecksum, newChecksum):
		return ChangeReasonSourceAndConfig
	case configChecksumChanged(oldChecksum, newChecksum):
		return ChangeReasonConfigChanged
	default:
		return ChangeReasonSourceChanged
	}
}
func reuseExistingBuild(id string) *CheckBuildNeededOutput {
	return &CheckBuildNeededOutput{ExistingBuildID: id, ChangeReason: ChangeReasonNoChanges}
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) CheckBuildNeeded(ctx context.Context, input *CheckBuildNeededInput) (*CheckBuildNeededOutput, error) {
	var cfg app.ComponentConfigConnection
	if err := a.db.WithContext(ctx).Scopes(buildsource.PreloadConnection).Where(app.ComponentConfigConnection{AppConfigID: input.NewAppConfigID, ComponentID: input.ComponentID}).First(&cfg).Error; err != nil {
		return nil, err
	}
	out := &CheckBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}
	if input.OldAppConfigID == "" {
		return out, nil
	}
	var old app.ComponentConfigConnection
	if err := a.db.WithContext(ctx).Where(app.ComponentConfigConnection{AppConfigID: input.OldAppConfigID, ComponentID: input.ComponentID}).First(&old).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, nil
		}
		return nil, err
	}
	existing, err := buildselection.Component(ctx, a.db, old.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return out, nil
	}
	var completed app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Preload("ComponentConfigConnection", buildsource.PreloadConnection).
		Where(app.ComponentBuild{ID: existing.ID}).First(&completed).Error; err != nil {
		return nil, err
	}
	if completed.Status != app.ComponentBuildStatusActive {
		return out, nil
	}
	if cfg.Checksum == "" || completed.ComponentConfigConnection.Checksum == "" || cfg.Checksum != completed.ComponentConfigConnection.Checksum {
		out.ChangeReason = buildChangeReason(sourceCommitChanged(input), completed.ComponentConfigConnection.Checksum, cfg.Checksum)
		return out, nil
	}
	oldChecksum, err := buildsource.ComponentChecksum(&completed.ComponentConfigConnection)
	if err != nil {
		return nil, err
	}
	newChecksum, err := buildsource.ComponentChecksum(&cfg)
	if err != nil {
		return nil, err
	}
	out.ChangeReason = buildChangeReason(sourceCommitChanged(input), oldChecksum, newChecksum)
	if oldChecksum != newChecksum {
		return out, nil
	}
	candidate := app.ComponentBuild{ComponentConfigConnection: cfg}
	if input.RunID != "" {
		candidate.AppBranchRunID = &input.RunID
	}
	identity, err := buildsource.ComponentSelection(ctx, a.db, a.vcsHelpers, &candidate)
	if err != nil {
		return nil, err
	}
	equal, err := buildsource.SourcesEqual(ctx, a.vcsHelpers, cfg.ConnectedGithubVCSConfig, cfg.PublicGitVCSConfig, completed.SourceIdentity, identity)
	if err != nil {
		return nil, err
	}
	if !equal {
		return out, nil
	}
	return reuseExistingBuild(existing.ID), nil
}
