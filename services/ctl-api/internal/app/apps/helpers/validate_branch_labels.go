package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// LatestConfigInstallGroups returns the install groups on a branch's newest
// config, or nothing when the branch has no config yet.
func (h *Helpers) LatestConfigInstallGroups(ctx context.Context, branchID string) ([]app.AppBranchInstallGroup, error) {
	return LatestConfigInstallGroupsWithDB(ctx, h.db, branchID)
}

func LatestConfigInstallGroupsWithDB(ctx context.Context, db *gorm.DB, branchID string) ([]app.AppBranchInstallGroup, error) {
	var latestConfig app.AppBranchConfig
	if err := db.WithContext(ctx).
		Where(app.AppBranchConfig{AppBranchID: branchID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest config for branch %s: %w", branchID, err)
	}

	var groups []app.AppBranchInstallGroup
	if err := db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: latestConfig.ID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to get install groups for config %s: %w", latestConfig.ID, err)
	}

	return groups, nil
}
