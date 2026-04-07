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

	// Check if a commit with this SHA already exists for the same owner+branch
	var existing app.VCSConnectionCommit
	findRes := a.db.WithContext(ctx).
		Where("sha = ? AND owner_id = ? AND owner_type = ? AND branch = ?",
			vcsCommit.SHA, vcsCommit.OwnerID, vcsCommit.OwnerType, vcsCommit.Branch).
		First(&existing)
	if findRes.Error == nil {
		return &existing, nil
	}

	createRes := a.db.WithContext(ctx).Create(vcsCommit)
	if createRes.Error != nil {
		return nil, fmt.Errorf("unable to create VCS commit record: %w", createRes.Error)
	}

	return vcsCommit, nil
}
