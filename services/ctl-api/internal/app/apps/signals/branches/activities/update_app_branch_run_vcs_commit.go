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
// @by-field runID
func (a *Activities) updateAppBranchRunVCSCommit(ctx context.Context, runID, vcsCommitID string) error {
	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where("id = ?", runID).
		Update("vcs_connection_commit_id", vcsCommitID)

	if res.Error != nil {
		return fmt.Errorf("unable to update run VCS commit: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("app branch run not found: %s: %w", runID, gorm.ErrRecordNotFound)
	}

	return nil
}
