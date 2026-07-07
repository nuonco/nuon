package activities

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
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
// @start-to-close-timeout 30s
func (a *Activities) CreateOrUpdatePRComment(ctx context.Context, input *CreateOrUpdatePRCommentInput) (*CreateOrUpdatePRCommentOutput, error) {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping PR comment: no authenticated GitHub client available")
		return &CreateOrUpdatePRCommentOutput{}, nil
	}

	comment := &github.IssueComment{
		Body: &input.Body,
	}

	if input.ExistingCommentID != nil && *input.ExistingCommentID != 0 {
		updated, _, err := client.Issues.EditComment(ctx, owner, repo, *input.ExistingCommentID, comment)
		if err != nil {
			if nrErr := nonRetryableGitHubError(err); nrErr != nil {
				return nil, nrErr
			}
			return nil, fmt.Errorf("unable to edit PR comment: %w", err)
		}
		return &CreateOrUpdatePRCommentOutput{CommentID: updated.GetID()}, nil
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
