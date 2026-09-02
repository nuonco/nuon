package activities

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
)

const maxChangedFilePaths = 3000
const maxComparedFilePaths = 300

type FetchChangedFilePathsInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
	PRNumber    *int   `json:"pr_number,omitempty"`
	BaseSHA     string `json:"base_sha,omitempty"`

	// BaseBranch compares the commit against a branch tip, which is what a
	// pull request run wants. Empty asks for the commit's own diff, which is
	// what a push run wants.
	BaseBranch string `json:"base_branch,omitempty"`
}

type ChangedFilePaths struct {
	Paths []string `json:"paths"`

	// Truncated reports that the diff exceeded maxChangedFilePaths, so Paths is
	// only a prefix and no caller may conclude the whole diff is ignorable.
	Truncated bool `json:"truncated,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 60s
func (a *Activities) FetchChangedFilePaths(ctx context.Context, input *FetchChangedFilePathsInput) (*ChangedFilePaths, error) {
	owner, repo, client, err := a.resolveGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		return nil, err
	}

	out := &ChangedFilePaths{}
	opts := &github.ListOptions{PerPage: 100}

	for {
		var (
			files    []*github.CommitFile
			nextPage int
		)

		if input.PRNumber != nil {
			prFiles, resp, listErr := client.PullRequests.ListFiles(ctx, owner, repo, *input.PRNumber, opts)
			if listErr != nil {
				if nrErr := nonRetryableGitHubError(listErr); nrErr != nil {
					return nil, nrErr
				}
				return nil, fmt.Errorf("unable to list pull request files: %w", listErr)
			}
			files = prFiles
			if resp != nil {
				nextPage = resp.NextPage
			}
		} else if input.BaseSHA != "" || input.BaseBranch != "" {
			baseRef := input.BaseSHA
			if baseRef == "" {
				baseRef = input.BaseBranch
			}
			comparison, resp, cmpErr := client.Repositories.CompareCommits(ctx, owner, repo, baseRef, input.CommitSHA, opts)
			if cmpErr != nil {
				if nrErr := nonRetryableGitHubError(cmpErr); nrErr != nil {
					return nil, nrErr
				}
				return nil, fmt.Errorf("unable to compare commits: %w", cmpErr)
			}
			files = comparison.Files
			if resp != nil {
				nextPage = resp.NextPage
			}
		} else {
			commit, resp, cmtErr := client.Repositories.GetCommit(ctx, owner, repo, input.CommitSHA, opts)
			if cmtErr != nil {
				if nrErr := nonRetryableGitHubError(cmtErr); nrErr != nil {
					return nil, nrErr
				}
				return nil, fmt.Errorf("unable to get commit: %w", cmtErr)
			}
			files = commit.Files
			if resp != nil {
				nextPage = resp.NextPage
			}
		}

		for _, f := range files {
			if path := f.GetFilename(); path != "" {
				out.Paths = append(out.Paths, path)
			}
			if len(out.Paths) >= maxChangedFilePaths {
				out.Truncated = true
				return out, nil
			}
		}
		if input.PRNumber == nil && len(out.Paths) >= maxComparedFilePaths {
			out.Truncated = true
			return out, nil
		}

		if nextPage == 0 {
			break
		}
		opts.Page = nextPage
	}

	return out, nil
}
