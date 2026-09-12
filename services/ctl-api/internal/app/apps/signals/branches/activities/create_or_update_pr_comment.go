package activities

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
)

const (
	previewCommentsPerPage = 100
	previewCommentsMaxPage = 10
)

type CreateOrUpdatePRCommentInput struct {
	VcsConfigID       string `json:"vcs_config_id" validate:"required"`
	PRNumber          int    `json:"pr_number" validate:"required"`
	ExistingCommentID *int64 `json:"existing_comment_id,omitempty"`
	Body              string `json:"body" validate:"required"`
}

type CreateOrUpdatePRCommentOutput struct {
	CommentID int64 `json:"comment_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) CreateOrUpdatePRComment(ctx context.Context, input *CreateOrUpdatePRCommentInput) (*CreateOrUpdatePRCommentOutput, error) {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping PR comment: no authenticated GitHub client available")
		return &CreateOrUpdatePRCommentOutput{}, nil
	}

	var priorComments []*github.IssueComment
	if marker, ok := ParsePreviewCommentMarker(input.Body); ok {
		priorComments, err = a.listPreviewComments(ctx, client, owner, repo, input.PRNumber, marker.Name)
		if err != nil {
			a.l.Warn("unable to list previous preview comments",
				zap.Int("pr_number", input.PRNumber),
				zap.Error(err))
		}
	}

	targetCommentID := int64(0)
	if input.ExistingCommentID != nil {
		targetCommentID = *input.ExistingCommentID
	}
	if targetCommentID == 0 {
		// A signal that never learned the run's comment ID — or a retry after the
		// activity timed out mid-write — would otherwise stack another report onto
		// the PR, so reuse the newest live one for this preview instead.
		targetCommentID = newestLivePreviewCommentID(priorComments)
	}

	commentID, err := a.writePRComment(ctx, client, owner, repo, input.PRNumber, targetCommentID, input.Body)
	if err != nil {
		return nil, err
	}

	a.collapseSupersededPreviewComments(ctx, client, owner, repo, priorComments, commentID)

	return &CreateOrUpdatePRCommentOutput{CommentID: commentID}, nil
}

func (a *Activities) writePRComment(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	prNumber int,
	commentID int64,
	body string,
) (int64, error) {
	comment := &github.IssueComment{Body: &body}

	if commentID != 0 {
		updated, _, err := client.Issues.EditComment(ctx, owner, repo, commentID, comment)
		if err != nil {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return 0, nrErr
			}
			return 0, fmt.Errorf("unable to edit PR comment: %w", err)
		}
		return updated.GetID(), nil
	}

	created, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return 0, nrErr
		}
		return 0, fmt.Errorf("unable to create PR comment: %w", err)
	}
	return created.GetID(), nil
}

// listPreviewComments returns the PR's comments that are reports on the named
// preview, oldest first.
func (a *Activities) listPreviewComments(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	prNumber int,
	previewName string,
) ([]*github.IssueComment, error) {
	if previewName == "" {
		return nil, nil
	}

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: previewCommentsPerPage},
	}

	var matches []*github.IssueComment
	for page := 0; page < previewCommentsMaxPage; page++ {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return matches, nrErr
			}
			return matches, fmt.Errorf("unable to list PR comments: %w", err)
		}

		for _, comment := range comments {
			marker, ok := ParsePreviewCommentMarker(comment.GetBody())
			if ok && marker.Name == previewName {
				matches = append(matches, comment)
			}
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return matches, nil
}

func newestLivePreviewCommentID(comments []*github.IssueComment) int64 {
	for i := len(comments) - 1; i >= 0; i-- {
		if !IsCollapsedPreviewComment(comments[i].GetBody()) {
			return comments[i].GetID()
		}
	}
	return 0
}

// collapseSupersededPreviewComments folds every earlier report on the same
// preview into a collapsed <details> block, so a PR that has been through many
// iterations shows only the current report expanded. Failures are logged and
// swallowed: the report itself is already posted and collapsing is cosmetic.
func (a *Activities) collapseSupersededPreviewComments(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	comments []*github.IssueComment,
	currentCommentID int64,
) {
	for _, comment := range comments {
		if comment.GetID() == currentCommentID {
			continue
		}

		collapsed, ok := CollapsePreviewCommentBody(comment.GetBody())
		if !ok {
			continue
		}

		if _, _, err := client.Issues.EditComment(ctx, owner, repo, comment.GetID(),
			&github.IssueComment{Body: &collapsed}); err != nil {
			a.l.Warn("unable to collapse superseded preview comment",
				zap.Int64("comment_id", comment.GetID()),
				zap.Error(err))
		}
	}
}
