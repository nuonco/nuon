package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type FetchCommitBySHAInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	SHA         string `json:"sha" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) FetchCommitBySHA(ctx context.Context, input *FetchCommitBySHAInput) (*app.VCSConnectionCommit, error) {
	vcsHelpers := a.helpers.VCSHelpers()

	var connectedCfg app.ConnectedGithubVCSConfig
	connectedRes := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", input.VcsConfigID)

	if connectedRes.Error == nil {
		client, err := vcsHelpers.GetVCSConnectionClient(ctx, &connectedCfg.VCSConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to get VCS client: %w", err)
		}

		ghCommit, _, err := client.Repositories.GetCommit(ctx, connectedCfg.RepoOwner, connectedCfg.RepoName, input.SHA, nil)
		if err != nil {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return nil, nrErr
			}
			return nil, fmt.Errorf("unable to get commit %s: %w", input.SHA, err)
		}

		vcsCommit := vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit,
			connectedCfg.ID,
			plugins.TableName(a.db, connectedCfg),
			connectedCfg.VCSConnectionID)
		if vcsCommit == nil {
			return nil, fmt.Errorf("invalid commit data from GitHub for SHA %s", input.SHA)
		}

		return vcsCommit, nil
	}

	var publicCfg app.PublicGitVCSConfig
	publicRes := a.db.WithContext(ctx).First(&publicCfg, "id = ?", input.VcsConfigID)
	if publicRes.Error == nil {
		owner, repo, client, err := a.resolveGithubClient(ctx, input.VcsConfigID)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve github client: %w", err)
		}

		ghCommit, _, err := client.Repositories.GetCommit(ctx, owner, repo, input.SHA, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to get commit %s: %w", input.SHA, err)
		}

		vcsCommit := vcsHelpers.GithubCommitToVCSConnectionCommit(ghCommit,
			publicCfg.ID,
			plugins.TableName(a.db, publicCfg),
			"")
		if vcsCommit == nil {
			return nil, fmt.Errorf("invalid commit data from GitHub for SHA %s", input.SHA)
		}

		return vcsCommit, nil
	}

	return nil, fmt.Errorf("VCS config not found: %s", input.VcsConfigID)
}
