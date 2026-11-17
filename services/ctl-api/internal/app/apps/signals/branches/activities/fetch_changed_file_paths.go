package activities

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/go-github/v50/github"
	"go.temporal.io/sdk/temporal"
)

const maxChangedFilePaths = 3000
const maxComparedFilePaths = 300

var commitSHARegexp = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func isCommitSHA(s string) bool {
	return commitSHARegexp.MatchString(s)
}

func isZeroCommitSHA(s string) bool {
	return s != "" && strings.Trim(s, "0") == ""
}

// isSafeGitRef rejects refs that could change the GitHub API URL they are
// interpolated into.
func isSafeGitRef(ref string) bool {
	if ref == "" || strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") ||
		strings.Contains(ref, "..") || strings.ContainsAny(ref, `\?#%`) {
		return false
	}
	for _, r := range ref {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func validateChangedFilePathsInput(input *FetchChangedFilePathsInput) error {
	if !isCommitSHA(input.CommitSHA) {
		return fmt.Errorf("commit_sha must be a full commit SHA")
	}
	if input.BaseSHA != "" && !isCommitSHA(input.BaseSHA) {
		return fmt.Errorf("base_sha must be a full commit SHA")
	}
	if input.BaseBranch != "" && !isSafeGitRef(input.BaseBranch) {
		return fmt.Errorf("base_branch is not a valid git ref")
	}
	if input.PRNumber != nil && *input.PRNumber <= 0 {
		return fmt.Errorf("pr_number must be positive")
	}
	return nil
}

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
	if err := validateChangedFilePathsInput(input); err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "InvalidChangedFilePathsInput", err)
	}

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
