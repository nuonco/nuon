package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appBranchID
func (a *Activities) updateAppBranchLastSyncedCommit(ctx context.Context, appBranchID, commitSHA string) error {
	res := a.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Where("id = ?", appBranchID).
		Update("last_synced_commit", commitSHA)

	if res.Error != nil {
		return fmt.Errorf("unable to update last synced commit: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("app branch not found: %s", appBranchID)
	}

	return nil
}
