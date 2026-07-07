package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FindLatestNonPreviewAppConfigInput struct {
	AppID string `json:"app_id" validate:"required"`
}

type FindLatestNonPreviewAppConfigOutput struct {
	AppConfigID string `json:"app_config_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) FindLatestNonPreviewAppConfig(ctx context.Context, input *FindLatestNonPreviewAppConfigInput) (*FindLatestNonPreviewAppConfigOutput, error) {
	var cfg app.AppConfig
	err := a.db.WithContext(ctx).
		Where(app.AppConfig{
			AppID:  input.AppID,
			Status: app.AppConfigStatusActive,
		}).
		Where("labels->>'source' IS NULL OR labels->>'source' != ?", string(app.AppBranchRunTypeGitPreview)).
		Where("intermediate_config IS NOT NULL").
		Order("created_at DESC").
		First(&cfg).Error
	if err == nil {
		return &FindLatestNonPreviewAppConfigOutput{AppConfigID: cfg.ID}, nil
	}

	// Fallback: use the latest config with an intermediate config (may be a previous preview).
	// This lets subsequent PR pushes diff against each other.
	err = a.db.WithContext(ctx).
		Where(app.AppConfig{
			AppID:  input.AppID,
			Status: app.AppConfigStatusActive,
		}).
		Where("intermediate_config IS NOT NULL").
		Order("created_at DESC").
		First(&cfg).Error
	if err != nil {
		return &FindLatestNonPreviewAppConfigOutput{}, nil
	}

	return &FindLatestNonPreviewAppConfigOutput{AppConfigID: cfg.ID}, nil
}
