package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appBranchID
func (a *Activities) getLatestAppBranchRunCommitSHA(ctx context.Context, appBranchID string) (string, error) {
	var run app.AppBranchRun
	res := a.db.WithContext(ctx).
		Where("app_branch_id = ? AND status = ? AND commit_sha != ''", appBranchID, "success").
		Order("created_at DESC").
		First(&run)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "", nil // No previous successful run with a commit
		}
		return "", fmt.Errorf("unable to get latest run commit: %w", res.Error)
	}
	return run.CommitSHA, nil
}
