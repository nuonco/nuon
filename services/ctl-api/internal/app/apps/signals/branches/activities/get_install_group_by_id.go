package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetInstallGroupByID(ctx context.Context, groupID string) (*app.AppBranchInstallGroup, error) {
	var group app.AppBranchInstallGroup

	err := a.db.WithContext(ctx).
		Where("id = ?", groupID).
		First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrapf(err, "install group not found: %s", groupID)
		}
		return nil, errors.Wrap(err, "unable to query install group")
	}

	return &group, nil
}
