package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	githubpkg "github.com/nuonco/nuon/pkg/github/repo"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	token, err := h.CreateInstallationToken(ctx, &cfg.VCSConnection, cfg.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create installation token")
	}

	commit, err := h.GetConnectedGithubVCSConfigLatestCommit(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest commit")
	}

	return &plantypes.GitSource{
		URL:  githubpkg.RepoPath(cfg.RepoOwner, cfg.RepoName, token),
		Ref:  generics.FromPtrStr(commit.SHA),
		Path: cfg.Directory,
	}, nil
}

// GetGitSourceAtCommit returns a git source for a connected GitHub repo at a specific commit SHA.
// Unlike GetGitSource, it does not look up the latest commit — it uses the provided SHA directly.
func (h *Helpers) GetGitSourceAtCommit(ctx context.Context, cfg *app.ConnectedGithubVCSConfig, commitSHA string) (*plantypes.GitSource, error) {
	token, err := h.CreateInstallationToken(ctx, &cfg.VCSConnection, cfg.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create installation token")
	}

	return &plantypes.GitSource{
		URL:  githubpkg.RepoPath(cfg.RepoOwner, cfg.RepoName, token),
		Ref:  commitSHA,
		Path: cfg.Directory,
	}, nil
}

// GetPublicGitSourceAtCommit returns a git source for a public repo at a specific commit SHA.
func (h *Helpers) GetPublicGitSourceAtCommit(cfg *app.PublicGitVCSConfig, commitSHA string) (*plantypes.GitSource, error) {
	url, err := githubpkg.EnsureURL(cfg.Repo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to derive url from source")
	}

	return &plantypes.GitSource{
		URL:  url,
		Ref:  commitSHA,
		Path: cfg.Directory,
	}, nil
}

// CreateInstallationToken creates a GitHub installation token for the given VCS connection and repo.
func (h *Helpers) CreateInstallationToken(ctx context.Context, vcsConn *app.VCSConnection, repoName string) (string, error) {
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
