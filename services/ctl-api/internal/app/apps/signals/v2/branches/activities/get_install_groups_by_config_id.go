package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen activity
func (a *Activities) GetInstallGroupsByConfigID(ctx context.Context, configID string) ([]*app.AppBranchInstallGroup, error) {
	var groups []*app.AppBranchInstallGroup

	// Query install groups for this config, ordered by the order field
	err := a.db.WithContext(ctx).
		Where("app_branch_config_id = ?", configID).
		Order("\"order\" ASC").
		Find(&groups).Error
	if err != nil {
		return nil, errors.Wrap(err, "unable to query install groups")
	}

	return groups, nil
}
