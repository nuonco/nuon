package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) updateAppBranchRunCommitSHA(ctx context.Context, runID, commitSHA string) error {
	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where("id = ?", runID).
		Update("commit_sha", commitSHA)

	if res.Error != nil {
		return fmt.Errorf("unable to update run commit SHA: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("app branch run not found: %s", runID)
	}

	return nil
}
