package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PreviewSourcePR struct {
	PRNumber        int                    `json:"pr_number"`
	Title           string                 `json:"title"`
	HeadSHA         string                 `json:"head_sha"`
	HeadRef         string                 `json:"head_ref"`
	URL             string                 `json:"url"`
	Author          string                 `json:"author,omitempty"`
	AuthorAvatarURL string                 `json:"author_avatar_url,omitempty"`
	IsDraft         bool                   `json:"is_draft"`
	Labels          []PreviewSourcePRLabel `json:"labels"`
	CreatedAt       *time.Time             `json:"created_at,omitempty"`
	UpdatedAt       *time.Time             `json:"updated_at,omitempty"`
}

type PreviewSourcePRLabel struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type PreviewSourceBranch struct {
	Name string `json:"name"`
	SHA  string `json:"sha,omitempty"`
}

type ListPreviewSourcesResult struct {
	PullRequests []PreviewSourcePR     `json:"pull_requests"`
	Branches     []PreviewSourceBranch `json:"branches"`
}

func (h *Helpers) ListPreviewSources(ctx context.Context, branch *app.AppBranch, config *app.AppBranchConfig) (*ListPreviewSourcesResult, error) {
	result := &ListPreviewSourcesResult{
		PullRequests: []PreviewSourcePR{},
		Branches:     []PreviewSourceBranch{},
	}
	preview := branchPreviewConfigOrDefault(config)

	owner, repo, client, err := h.resolveGithubClientForBranchConfig(ctx, config)
	if err != nil {
		return result, err
	}

	gitBase := previewGitBase(branch, config)

	prOpts := &github.PullRequestListOptions{
		Base:  gitBase,
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		prs, resp, err := client.PullRequests.List(ctx, owner, repo, prOpts)
		if err != nil {
			return nil, fmt.Errorf("unable to list pull requests: %w", err)
		}

		for _, pr := range prs {
			if pr.GetDraft() && preview.IgnoreDrafts {
				continue
			}

			headRef := ""
			headSHA := ""
			if pr.Head != nil {
				headRef = pr.Head.GetRef()
				headSHA = pr.Head.GetSHA()
			}
			labels := make([]PreviewSourcePRLabel, 0, len(pr.Labels))
			for _, label := range pr.Labels {
				name := label.GetName()
				if name == "" {
					continue
				}
				labels = append(labels, PreviewSourcePRLabel{
					Name:  name,
					Color: label.GetColor(),
				})
			}
			entry := PreviewSourcePR{
				PRNumber:        pr.GetNumber(),
				Title:           pr.GetTitle(),
				HeadSHA:         headSHA,
				HeadRef:         headRef,
				URL:             pr.GetHTMLURL(),
				Author:          pr.GetUser().GetLogin(),
				AuthorAvatarURL: pr.GetUser().GetAvatarURL(),
				IsDraft:         pr.GetDraft(),
				Labels:          labels,
			}
			if pr.CreatedAt != nil {
				created := pr.CreatedAt.Time
				entry.CreatedAt = &created
			}
			if pr.UpdatedAt != nil {
				updated := pr.UpdatedAt.Time
				entry.UpdatedAt = &updated
			}
			result.PullRequests = append(result.PullRequests, entry)
		}

		if resp.NextPage == 0 {
			break
		}
		prOpts.Page = resp.NextPage
	}

	branchOpts := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, branchOpts)
		if err != nil {
			return nil, fmt.Errorf("unable to list branches: %w", err)
		}

		for _, b := range branches {
			name := b.GetName()
			if name == gitBase {
				continue
			}
			entry := PreviewSourceBranch{Name: name}
			if b.Commit != nil {
				entry.SHA = b.Commit.GetSHA()
			}
			result.Branches = append(result.Branches, entry)
		}

		if resp.NextPage == 0 {
			break
		}
		branchOpts.Page = resp.NextPage
	}

	return result, nil
}

func previewGitBase(branch *app.AppBranch, config *app.AppBranchConfig) string {
	if config != nil {
		if cfg := config.ConnectedGithubVCSConfig; cfg != nil && cfg.Branch != "" {
			return cfg.Branch
		}
		if cfg := config.PublicGitVCSConfig; cfg != nil && cfg.Branch != "" {
			return cfg.Branch
		}
	}
	if branch != nil {
		return branch.Name
	}
	return ""
}

func (h *Helpers) resolveGithubClientForBranchConfig(ctx context.Context, config *app.AppBranchConfig) (owner, repo string, client *github.Client, err error) {
	switch {
	case config.ConnectedGithubVCSConfig != nil:
		cfg := config.ConnectedGithubVCSConfig
		client, err = h.vcsHelpers.GetVCSConnectionClient(ctx, &cfg.VCSConnection)
		if err != nil {
			return "", "", nil, fmt.Errorf("unable to get VCS client: %w", err)
		}
		return cfg.RepoOwner, cfg.RepoName, client, nil
	case config.PublicGitVCSConfig != nil:
		cfg := config.PublicGitVCSConfig
		repoOwner, repoName, parseErr := parseGithubRepoURL(cfg.Repo)
		if parseErr != nil {
			return "", "", nil, parseErr
		}
		client, _, err = h.vcsHelpers.ResolvePublicRepoGithubClient(ctx, h.l, cfg.OrgID, repoOwner)
		if err != nil {
			return "", "", nil, err
		}
		return repoOwner, repoName, client, nil
	default:
		return "", "", nil, fmt.Errorf("branch config has no VCS configuration")
	}
}

func parseGithubRepoURL(repoURL string) (owner, name string, err error) {
	raw := repoURL
	raw = strings.TrimPrefix(raw, "https://github.com/")
	raw = strings.TrimPrefix(raw, "http://github.com/")
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimSuffix(raw, ".git")

	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unable to parse GitHub repo: %s", repoURL)
	}
	return parts[0], parts[1], nil
}
