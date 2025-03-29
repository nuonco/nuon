package helpers

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) GetVCSConfigLatestCommit(ctx context.Context, vcsCfg *app.ConnectedGithubVCSConfig) (*github.RepositoryCommit, error) {
	client, err := h.GetVCSConnectionClient(ctx, &vcsCfg.VCSConnection)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         err,
			Description: "invalid VCS connection, unable to get access token",
		}
	}

	commitResp, _, err := client.Repositories.GetCommit(ctx, vcsCfg.RepoOwner, vcsCfg.RepoName, vcsCfg.Branch, &github.ListOptions{})
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("unable to get latest commit: %w", err),
			Description: "unable to get error",
		}
	}

	return commitResp, nil
}
