package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/pkg/generics"
	githubpkg "github.com/nuonco/nuon/pkg/github/repo"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) GetConnectedGithubVCSConfigLatestCommit(ctx context.Context, vcsCfg *app.ConnectedGithubVCSConfig) (*github.RepositoryCommit, error) {
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

func (h *Helpers) GetPublicGitVCSConfigLatestCommit(ctx context.Context, cfg *app.PublicGitVCSConfig) (*github.RepositoryCommit, error) {
	owner, repo, err := parseOwnerRepo(cfg.Repo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse repo %q: %w", cfg.Repo, err)
	}

	// Use an unauthenticated client for public repos
	client := github.NewClient(nil)

	commitResp, _, err := client.Repositories.GetCommit(ctx, owner, repo, cfg.Branch, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get latest commit for %s/%s@%s: %w", owner, repo, cfg.Branch, err)
	}

	return commitResp, nil
}

// parseOwnerRepo extracts the GitHub owner and repo name from either
// an "owner/repo" string or a "https://github.com/owner/repo" URL.
func parseOwnerRepo(repoStr string) (string, string, error) {
	// Try owner/repo format first
	owner, repo, err := githubpkg.ParseRepo(repoStr)
	if err == nil {
		return owner, repo, nil
	}

	// Try URL format: https://github.com/owner/repo[.git]
	trimmed := strings.TrimPrefix(repoStr, "https://github.com/")
	if trimmed == repoStr {
		return "", "", fmt.Errorf("unsupported repo format: %s", repoStr)
	}
	trimmed = strings.TrimSuffix(trimmed, ".git")
	trimmed = strings.TrimSuffix(trimmed, "/")

	return githubpkg.ParseRepo(trimmed)
}

// GithubCommitToVCSConnectionCommit converts a GitHub API commit response to a VCSConnectionCommit struct.
// The struct is not saved to the database - caller is responsible for persistence.
func (h *Helpers) GithubCommitToVCSConnectionCommit(ghCommit *github.RepositoryCommit, ownerID, ownerType string, vcsConnectionID string) *app.VCSConnectionCommit {
	if ghCommit == nil || ghCommit.SHA == nil {
		return nil
	}

	vcsCommit := &app.VCSConnectionCommit{
		SHA:       *ghCommit.SHA,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	}
	if vcsConnectionID != "" {
		vcsCommit.VCSConnectionID = generics.ToPtr(vcsConnectionID)
	}

	// Populate commit metadata with defensive nil checking
	if ghCommit.Commit != nil {
		if ghCommit.Commit.Author != nil {
			vcsCommit.AuthorName = generics.FromPtrStr(ghCommit.Commit.Author.Name)
			vcsCommit.AuthorEmail = generics.FromPtrStr(ghCommit.Commit.Author.Email)
		}
		vcsCommit.Message = generics.FromPtrStr(ghCommit.Commit.Message)
	}

	return vcsCommit
}
