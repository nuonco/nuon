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
	vcsHelpers := a.helpers.VCSHelpers()

	// Try ConnectedGithubVCSConfig first
	var connectedCfg app.ConnectedGithubVCSConfig
	res := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", vcsConfigID)
	if res.Error == nil {
		commit, err := vcsHelpers.GetConnectedGithubVCSConfigLatestCommit(ctx, &connectedCfg)
		if err != nil {
			return "", fmt.Errorf("unable to get latest commit for connected repo: %w", err)
		}

		return *commit.SHA, nil
	}

	// Try PublicGitVCSConfig
	var publicCfg app.PublicGitVCSConfig
	res = a.db.WithContext(ctx).First(&publicCfg, "id = ?", vcsConfigID)
	if res.Error == nil {
		commit, err := vcsHelpers.GetPublicGitVCSConfigLatestCommit(ctx, &publicCfg)
		if err != nil {
			return "", fmt.Errorf("unable to get latest commit for public repo: %w", err)
		}
		return *commit.SHA, nil
	}

	return "", fmt.Errorf("VCS config not found: %s", vcsConfigID)
}
