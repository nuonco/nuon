package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix AppBranches
// @by-field vcsCommit
func (a *Activities) createCommit(ctx context.Context, vcsCommit *app.VCSConnectionCommit) (*app.VCSConnectionCommit, error) {
	if vcsCommit == nil {
		return nil, fmt.Errorf("vcsCommit cannot be nil")
	}

	createRes := a.db.WithContext(ctx).Create(vcsCommit)
	if createRes.Error != nil {
		return nil, fmt.Errorf("unable to create VCS commit record: %w", createRes.Error)
	}

	return vcsCommit, nil
}
