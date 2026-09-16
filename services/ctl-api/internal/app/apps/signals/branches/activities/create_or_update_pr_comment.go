package activities

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
)

const prCommentMarkerPrefix = "<!-- nuon-app-branch-preview:"

type CreateOrUpdatePRCommentInput struct {
	VcsConfigID       string `json:"vcs_config_id" validate:"required"`
	PRNumber          int    `json:"pr_number" validate:"required"`
	AppBranchID       string `json:"app_branch_id,omitempty"`
	ExistingCommentID *int64 `json:"existing_comment_id,omitempty"`
	Body              string `json:"body" validate:"required"`
}

type CreateOrUpdatePRCommentOutput struct {
	CommentID int64 `json:"comment_id"`
}

func PRCommentMarker(appBranchID string) string {
	if appBranchID == "" {
		return ""
	}
	return prCommentMarkerPrefix + appBranchID + " -->"
}

func commentHasMarker(body, appBranchID string) bool {
	marker := PRCommentMarker(appBranchID)
	return marker != "" && strings.Contains(body, marker)
}

func normalizePRCommentBody(body, appBranchID string, updatedAt time.Time) string {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	filtered := make([]string, 0, len(lines)+3)
	for _, line := range lines {
		if strings.HasPrefix(line, prCommentMarkerPrefix) || strings.HasPrefix(line, lastUpdatedPrefix) {
			continue
		}
		filtered = append(filtered, line)
	}

	headingIdx := -1
	for idx, line := range filtered {
		if strings.HasPrefix(line, "## ") {
			headingIdx = idx
			break
		}
	}
	if headingIdx >= 0 {
		withTimestamp := make([]string, 0, len(filtered)+2)
		withTimestamp = append(withTimestamp, filtered[:headingIdx+1]...)
		withTimestamp = append(withTimestamp, "", lastUpdatedLine(updatedAt))
		filtered = append(withTimestamp, filtered[headingIdx+1:]...)
	} else {
		filtered = append([]string{lastUpdatedLine(updatedAt), ""}, filtered...)
	}
	if marker := PRCommentMarker(appBranchID); marker != "" {
		filtered = append([]string{marker}, filtered...)
	}

	return strings.Join(filtered, "\n") + "\n"
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CreateOrUpdatePRComment(ctx context.Context, input *CreateOrUpdatePRCommentInput) (*CreateOrUpdatePRCommentOutput, error) {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping PR comment: no authenticated GitHub client available")
		return &CreateOrUpdatePRCommentOutput{}, nil
	}

	body := normalizePRCommentBody(input.Body, input.AppBranchID, time.Now())
	comment := &github.IssueComment{
		Body: &body,
	}

	return a.upsertPRComment(ctx, client, owner, repo, input, comment)
}

func (a *Activities) upsertPRComment(
	ctx context.Context,
	client *github.Client,
	owner string,
	repo string,
	input *CreateOrUpdatePRCommentInput,
	comment *github.IssueComment,
) (*CreateOrUpdatePRCommentOutput, error) {
	existingID := int64(0)
	if input.ExistingCommentID != nil {
		existingID = *input.ExistingCommentID
	}
	if existingID == 0 && input.AppBranchID != "" {
		if found, findErr := a.findPRCommentByMarker(ctx, client, owner, repo, input.PRNumber, input.AppBranchID); findErr != nil {
			a.l.Warn("unable to search for previous PR comment marker")
		} else {
			existingID = found
		}
	}

	if existingID != 0 {
		updated, _, err := client.Issues.EditComment(ctx, owner, repo, existingID, comment)
		if err == nil {
			return &CreateOrUpdatePRCommentOutput{CommentID: updated.GetID()}, nil
		}
		if !isGitHubNotFound(err) {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return nil, nrErr
			}
			return nil, fmt.Errorf("unable to edit PR comment: %w", err)
		}

		if input.AppBranchID != "" {
			found, findErr := a.findPRCommentByMarker(ctx, client, owner, repo, input.PRNumber, input.AppBranchID)
			if findErr != nil {
				a.l.Warn("unable to recover stale PR comment ID by marker")
			} else if found != 0 && found != existingID {
				updated, _, editErr := client.Issues.EditComment(ctx, owner, repo, found, comment)
				if editErr == nil {
					return &CreateOrUpdatePRCommentOutput{CommentID: updated.GetID()}, nil
				}
				if !isGitHubNotFound(editErr) {
					if nrErr := nonRetryableGitHubError(editErr); nrErr != nil {
						return nil, nrErr
					}
					return nil, fmt.Errorf("unable to edit recovered PR comment: %w", editErr)
				}
			}
		}
	}

	created, _, err := client.Issues.CreateComment(ctx, owner, repo, input.PRNumber, comment)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return nil, nrErr
		}
		return nil, fmt.Errorf("unable to create PR comment: %w", err)
	}

	return &CreateOrUpdatePRCommentOutput{CommentID: created.GetID()}, nil
}

func (a *Activities) findPRCommentByMarker(ctx context.Context, client *github.Client, owner, repo string, prNumber int, appBranchID string) (int64, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return 0, nrErr
			}
			return 0, err
		}
		for _, c := range comments {
			if c != nil && commentHasMarker(c.GetBody(), appBranchID) {
				return c.GetID(), nil
			}
		}
		if resp == nil || resp.NextPage == 0 {
			return 0, nil
		}
		opts.Page = resp.NextPage
	}
}
