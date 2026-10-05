package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type CheckSandboxBuildNeededInput struct {
	NewAppConfigID string `json:"new_app_config_id"`
	OldAppConfigID string `json:"old_app_config_id"`
	Force          bool   `json:"force"`
}

type CheckSandboxBuildNeededOutput struct {
	NeedsBuild      bool   `json:"needs_build"`
	ExistingBuildID string `json:"existing_build_id,omitempty"`
	ChangeReason    string `json:"change_reason,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CheckSandboxBuildNeeded(ctx context.Context, input *CheckSandboxBuildNeededInput) (*CheckSandboxBuildNeededOutput, error) {
	if input.Force || input.OldAppConfigID == "" {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	newCfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.NewAppConfigID)
	if err != nil {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	oldCfg, err := a.getAppSandboxConfigByAppConfigID(ctx, input.OldAppConfigID)
	if err != nil {
		return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
	}

	if !configdiff.SandboxConfigsEqual(*oldCfg, *newCfg) {
		return &CheckSandboxBuildNeededOutput{
			NeedsBuild:   true,
			ChangeReason: ChangeReasonConfigChanged,
		}, nil
	}

	var existing app.AppSandboxBuild
	err = a.db.WithContext(ctx).
		Where(app.AppSandboxBuild{
			AppConfigID: input.OldAppConfigID,
			Status:      app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&existing).Error
	if err == nil {
		return reuseExistingSandboxBuild(existing.ID), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to look up active sandbox build: %w", err)
	}

	var candidate app.AppSandboxBuild
	err = a.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  newCfg.AppID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&candidate).Error
	if err != nil {
		return nil, fmt.Errorf("unable to list active sandbox builds: %w", err)
	}
	if configdiff.SandboxConfigsEqual(candidate.AppSandboxConfig, *newCfg) {
		return reuseExistingSandboxBuild(candidate.ID), nil
	}

	return &CheckSandboxBuildNeededOutput{NeedsBuild: true, ChangeReason: ChangeReasonSourceChanged}, nil
}

func reuseExistingSandboxBuild(existingBuildID string) *CheckSandboxBuildNeededOutput {
	return &CheckSandboxBuildNeededOutput{
		NeedsBuild:      false,
		ExistingBuildID: existingBuildID,
		ChangeReason:    ChangeReasonNoChanges,
	}
}
