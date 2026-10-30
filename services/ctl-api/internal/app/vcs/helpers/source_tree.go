package helpers

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) GetSourceTree(ctx context.Context, orgID, repository, commit string, connection *app.VCSConnection) (*github.Tree, error) {
	if !sourceidentity.IsCommit(commit) {
		return nil, fmt.Errorf("source tree requires an immutable commit")
	}
	owner, repo, err := parseOwnerRepo(repository)
	if err != nil {
		return nil, err
	}
	scope := "public"
	if connection != nil {
		if connection.OrgID != orgID {
			return nil, fmt.Errorf("source connection does not belong to org")
		}
		scope = connection.ID
	}
	key := orgID + "\n" + scope + "\n" + owner + "/" + repo + "\n" + commit
	if h.sourceTrees != nil {
		if tree, ok := h.sourceTrees.Get(key); ok {
			return tree, nil
		}
	}
	result, err, _ := h.sourceFetches.Do(key, func() (any, error) {
		if h.sourceTrees != nil {
			if tree, ok := h.sourceTrees.Get(key); ok {
				return tree, nil
			}
		}
		var client *github.Client
		var err error
		if connection != nil {
			client, err = h.GetVCSConnectionClient(ctx, connection)
		} else {
			client, _, err = h.ResolvePublicRepoGithubClient(ctx, h.l, orgID, owner)
		}
		if err != nil {
			return nil, err
		}
		gitCommit, _, err := client.Git.GetCommit(ctx, owner, repo, commit)
		if err != nil {
			return nil, fmt.Errorf("unable to read pinned commit: %w", err)
		}
		if gitCommit == nil || gitCommit.Tree == nil || !sourceidentity.IsCommit(gitCommit.Tree.GetSHA()) {
			return nil, fmt.Errorf("pinned commit has no source tree")
		}
		tree, _, err := client.Git.GetTree(ctx, owner, repo, gitCommit.Tree.GetSHA(), true)
		if err != nil {
			return nil, fmt.Errorf("unable to read pinned source tree: %w", err)
		}
		if tree == nil || tree.GetTruncated() {
			return nil, fmt.Errorf("pinned source tree is missing or truncated")
		}
		if h.sourceTrees != nil {
			h.sourceTrees.Add(key, tree)
		}
		return tree, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*github.Tree), nil
}
