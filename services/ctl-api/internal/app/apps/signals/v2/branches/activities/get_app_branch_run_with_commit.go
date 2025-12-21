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
func (a *Activities) getAppBranchRunWithCommit(ctx context.Context, runID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	res := a.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		First(&run, "id = ?", runID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app branch run: %w", res.Error)
	}

	if run.VCSConnectionCommit == nil {
		return nil, fmt.Errorf("app branch run %s has no VCS connection commit", runID)
	}

	return &run, nil
}
