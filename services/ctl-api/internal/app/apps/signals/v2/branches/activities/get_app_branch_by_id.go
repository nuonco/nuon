package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix AppBranches
// @by-field appBranchID
func (a *Activities) getAppBranchByID(ctx context.Context, appBranchID string) (*app.AppBranch, error) {
	var branch app.AppBranch
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		Preload("Queue").
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_branch_configs_view_v1.created_at DESC").Limit(1)
		}).
		Preload("Configs.ConnectedGithubVCSConfig").
		Preload("Configs.PublicGitVCSConfig").
		Preload("Configs.InstallGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_branch_install_groups.order ASC")
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", res.Error)
	}

	return &branch, nil
}
