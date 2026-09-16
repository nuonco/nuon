package activities

import (
	"context"
	"fmt"
)

type CreatePRReactionInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	PRNumber    int    `json:"pr_number" validate:"required"`
	Content     string `json:"content,omitempty"`
}

type CreatePRReactionOutput struct {
	ReactionID int64 `json:"reaction_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CreatePRReaction(ctx context.Context, input *CreatePRReactionInput) (*CreatePRReactionOutput, error) {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping PR reaction: no authenticated GitHub client available")
		return &CreatePRReactionOutput{}, nil
	}

	content := input.Content
	if content == "" {
		content = "eyes"
	}

	reaction, _, err := client.Reactions.CreateIssueReaction(ctx, owner, repo, input.PRNumber, content)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return nil, nrErr
		}
		return nil, fmt.Errorf("unable to create PR reaction: %w", err)
	}
	if reaction == nil {
		return &CreatePRReactionOutput{}, nil
	}
	return &CreatePRReactionOutput{ReactionID: reaction.GetID()}, nil
}
