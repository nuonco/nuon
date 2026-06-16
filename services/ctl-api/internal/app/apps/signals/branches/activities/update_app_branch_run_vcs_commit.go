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
	// Look up the commit to get the SHA
	var commit app.VCSConnectionCommit
	if err := a.db.WithContext(ctx).First(&commit, "id = ?", vcsCommitID).Error; err != nil {
		return fmt.Errorf("unable to get VCS commit: %w", err)
	}

	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where("id = ?", runID).
		Updates(map[string]interface{}{
			"vcs_connection_commit_id": vcsCommitID,
			"commit_sha":               commit.SHA,
		})

	if res.Error != nil {
		return fmt.Errorf("unable to update run VCS commit: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("app branch run not found: %s: %w", runID, gorm.ErrRecordNotFound)
	}

	return nil
}
