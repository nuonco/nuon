package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FetchCommitPRInfoInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	Branch      string `json:"branch" validate:"required"`
}

type PRInfo struct {
	PRNumber      int    `json:"pr_number"`
	PRTitle       string `json:"pr_title"`
	PRStatus      string `json:"pr_status"`
	ReviewerCount int    `json:"pr_reviewer_count"`
	PRURL         string `json:"pr_url"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) FetchCommitPRInfo(ctx context.Context, input *FetchCommitPRInfoInput) (*PRInfo, error) {
	owner, repo, client, err := a.resolveGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		return nil, err
	}

	// List open PRs for this branch
	prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:  fmt.Sprintf("%s:%s", owner, input.Branch),
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list PRs: %w", err)
	}

	if len(prs) == 0 {
		return nil, nil
	}

	pr := prs[0]
	info := &PRInfo{
		PRNumber: pr.GetNumber(),
		PRTitle:  pr.GetTitle(),
		PRStatus: pr.GetState(),
		PRURL:    pr.GetHTMLURL(),
	}

	// Get reviewer count
	reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), &github.ListOptions{PerPage: 100})
	if err == nil {
		reviewers := map[string]bool{}
		for _, r := range reviews {
			if r.User != nil {
				reviewers[r.User.GetLogin()] = true
			}
		}
		info.ReviewerCount = len(reviewers)
	}

	// Also count requested reviewers
	if info.ReviewerCount == 0 {
		info.ReviewerCount = len(pr.RequestedReviewers)
	}

	return info, nil
}

type FetchCommitDiffStatsInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
	BaseBranch  string `json:"base_branch"`
}

type CommitDiffStats struct {
	FilesChanged int               `json:"files_changed"`
	Additions    int               `json:"additions"`
	Deletions    int               `json:"deletions"`
	ChangedFiles []ChangedFileInfo `json:"changed_files"`
}

type ChangedFileInfo struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) FetchCommitDiffStats(ctx context.Context, input *FetchCommitDiffStatsInput) (*CommitDiffStats, error) {
	owner, repo, client, err := a.resolveGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		return nil, err
	}

	base := input.BaseBranch
	if base == "" {
		base = "main"
	}

	comparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, base, input.CommitSHA, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("unable to compare commits: %w", err)
	}

	stats := &CommitDiffStats{
		FilesChanged: len(comparison.Files),
	}

	maxFiles := 6
	for i, f := range comparison.Files {
		stats.Additions += f.GetAdditions()
		stats.Deletions += f.GetDeletions()

		if i < maxFiles {
			stats.ChangedFiles = append(stats.ChangedFiles, ChangedFileInfo{
				Path:      f.GetFilename(),
				Additions: f.GetAdditions(),
				Deletions: f.GetDeletions(),
			})
		}
	}

	return stats, nil
}

// resolveGithubClient loads the VCS config and returns an authenticated GitHub client.
func (a *Activities) resolveGithubClient(ctx context.Context, vcsConfigID string) (owner, repo string, client *github.Client, err error) {
	vcsHelpers := a.helpers.VCSHelpers()

	// Try ConnectedGithubVCSConfig first
	var connectedCfg app.ConnectedGithubVCSConfig
	connectedRes := a.db.WithContext(ctx).
		Preload("VCSConnection").
		First(&connectedCfg, "id = ?", vcsConfigID)

	if connectedRes.Error == nil {
		client, err = vcsHelpers.GetVCSConnectionClient(ctx, &connectedCfg.VCSConnection)
		if err != nil {
			return "", "", nil, fmt.Errorf("unable to get VCS client: %w", err)
		}
		return connectedCfg.RepoOwner, connectedCfg.RepoName, client, nil
	}

	// Try PublicGitVCSConfig
	var publicCfg app.PublicGitVCSConfig
	publicRes := a.db.WithContext(ctx).First(&publicCfg, "id = ?", vcsConfigID)
	if publicRes.Error == nil {
		parts := strings.SplitN(strings.TrimPrefix(strings.TrimPrefix(publicCfg.Repo, "https://github.com/"), "/"), "/", 2)
		if len(parts) != 2 {
			return "", "", nil, fmt.Errorf("unable to parse public repo: %s", publicCfg.Repo)
		}
		return parts[0], strings.TrimSuffix(parts[1], ".git"), github.NewClient(nil), nil
	}

	return "", "", nil, fmt.Errorf("VCS config not found: %s", vcsConfigID)
}
