package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	githubpkg "github.com/powertoolsdev/mono/pkg/github/repo"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetPubliGitSource(ctx context.Context, cfg *app.PublicGitVCSConfig) (*plantypes.GitSource, error) {
	url, err := githubpkg.EnsureURL(cfg.Repo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to derive url from source")
	}

	return &plantypes.GitSource{
		URL:  url,
		Ref:  cfg.Branch,
		Path: cfg.Directory,
	}, nil
}

func (h *Helpers) GetGitSource(ctx context.Context, cfg *app.ConnectedGithubVCSConfig) (*plantypes.GitSource, error) {
	token, err := h.createInstallationToken(ctx, &cfg.VCSConnection, cfg.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create installation token")
	}

	commit, err := h.GetVCSConfigLatestCommit(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest commit")
	}

	return &plantypes.GitSource{
		URL:  githubpkg.RepoPath(cfg.RepoOwner, cfg.RepoName, token),
		Ref:  generics.FromPtrStr(commit.SHA),
		Path: cfg.Directory,
	}, nil
}

// NOTE(jm): this is mainly taken from `pkg/github` which was an implementation of this that used kube secrets to grab
// the GH app key and secret. Long term, this package should remove all need for that package.
func (h *Helpers) createInstallationToken(ctx context.Context, vcsConn *app.VCSConnection, repoName string) (string, error) {
	ghInstallID, err := strconv.Atoi(vcsConn.GithubInstallID)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse github install id")
	}

	resp, _, err := h.ghClient.Apps.CreateInstallationToken(ctx, int64(ghInstallID), &github.InstallationTokenOptions{
		Repositories: []string{repoName},
	})
	if err != nil {
		return "", fmt.Errorf("error creating installation token: %w", err)
	}

	if len(resp.Repositories) != 1 || *resp.Repositories[0].Name != repoName {
		return "", fmt.Errorf("installation does not allow accessing repo: %s", repoName)
	}

	return *resp.Token, nil
}
