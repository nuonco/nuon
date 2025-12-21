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
// @by-field vcsConfigID
func (a *Activities) getLatestCommitFromVCS(ctx context.Context, vcsConfigID string) (string, error) {
	// Fetch VCS config to get repo details
	var vcsConfig app.ConnectedGithubVCSConfig
	res := a.db.WithContext(ctx).First(&vcsConfig, "id = ?", vcsConfigID)
	if res.Error != nil {
		return "", fmt.Errorf("unable to get vcs config: %w", res.Error)
	}

	// TODO: Implement actual VCS integration
	// Use GitHub client via helpers to get latest commit
	// commit, err := a.helpers.GetLatestCommit(ctx, vcsConfig.Owner, vcsConfig.Repo, vcsConfig.Branch)

	return "", fmt.Errorf("GetLatestCommitFromVCS not yet implemented")
}
