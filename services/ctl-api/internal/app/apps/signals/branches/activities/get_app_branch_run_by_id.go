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
func (a *Activities) getAppBranchRunByID(ctx context.Context, runID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	res := a.db.WithContext(ctx).
		Preload("AppBranch").
		Preload("AppBranchConfig").
		Preload("AppBranchConfig.InstallGroups").
		Preload("Workflow").
		Preload("CreatedBy").
		// Optional: a run triggered with a pre-existing app config skips the
		// fetch-commit step and has none.
		Preload("VCSConnectionCommit").
		First(&run, "id = ?", runID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app branch run: %w", res.Error)
	}

	return &run, nil
}
