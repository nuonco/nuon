package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) getAppBranchRunWithCommit(ctx context.Context, runID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	res := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		First(&run, "id = ?", runID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app branch run: %w", res.Error)
	}

	// Non-retryable: a run either has a commit or never will, and retrying holds
	// the signal in-flight, which blocks every later run on the branch queue.
	if run.VCSConnectionCommit == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("app branch run %s has no VCS connection commit", runID),
			"APP_BRANCH_RUN_MISSING_COMMIT",
			nil,
		)
	}

	return &run, nil
}
