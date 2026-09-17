package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetLatestActiveBranchAppConfigInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
}

type GetLatestActiveBranchAppConfigOutput struct {
	AppConfigID    string `json:"app_config_id,omitempty"`
	AlreadyCurrent bool   `json:"already_current"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetLatestActiveBranchAppConfig(ctx context.Context, input *GetLatestActiveBranchAppConfigInput) (*GetLatestActiveBranchAppConfigOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var appConfig app.AppConfig
	err := a.db.WithContext(ctx).
		Where(app.AppConfig{
			AppID:       install.AppID,
			AppBranchID: generics.NewNullString(input.AppBranchID),
			Status:      app.AppConfigStatusActive,
		}).
		Where("labels->>'source' IS NULL OR labels->>'source' != ?", string(app.AppBranchRunTypeGitPreview)).
		Order("created_at DESC").
		First(&appConfig).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &GetLatestActiveBranchAppConfigOutput{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest active app config: %w", err)
	}

	return &GetLatestActiveBranchAppConfigOutput{
		AppConfigID:    appConfig.ID,
		AlreadyCurrent: install.AppConfigID == appConfig.ID,
	}, nil
}
