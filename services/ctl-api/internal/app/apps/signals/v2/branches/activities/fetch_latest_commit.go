package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix AppBranches
// @by-field vcsConfigID
func (a *Activities) fetchLatestCommit(ctx context.Context, vcsConfigID string) (*app.VCSConnectionCommit, error) {
	vcsHelpers := a.helpers.VCSHelpers()

	// Try ConnectedGithubVCSConfig first
	var connectedCfg app.ConnectedGithubVCSConfig
	connectedRes := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", vcsConfigID)

	if connectedRes.Error == nil {
		ghCommit, err := vcsHelpers.GetConnectedGithubVCSConfigLatestCommit(ctx, &connectedCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for connected repo: %w", err)
		}

		vcsCommit := vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit,
			connectedCfg.ID,
			plugins.TableName(a.db, connectedCfg),
			connectedCfg.VCSConnectionID)
		if vcsCommit == nil {
			return nil, fmt.Errorf("invalid commit data from GitHub")
		}

		return vcsCommit, nil
	}

	// Try PublicGitVCSConfig
	var publicCfg app.PublicGitVCSConfig
	publicRes := a.db.WithContext(ctx).First(&publicCfg, "id = ?", vcsConfigID)

	if publicRes.Error == nil {
		ghCommit, err := vcsHelpers.GetPublicGitVCSConfigLatestCommit(ctx, &publicCfg)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest commit for public repo: %w", err)
		}

		vcsCommit := vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit,
			publicCfg.ID,
			plugins.TableName(a.db, publicCfg),
			"")
		if vcsCommit == nil {
			return nil, fmt.Errorf("invalid commit data from GitHub")
		}

		return vcsCommit, nil
	}

	return nil, fmt.Errorf("VCS config not found: %s", vcsConfigID)
}
