package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field appBranchID
func (a *Activities) getAppBranchByID(ctx context.Context, appBranchID string) (*app.AppBranch, error) {
	var branch app.AppBranch
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		Preload("ConnectedGithubVCSConfig").
		Preload("Queue").
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", res.Error)
	}

	return &branch, nil
}
