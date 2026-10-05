package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
)

type ResolvePullRequestHeadSHAInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	PRNumber    int    `json:"pr_number" validate:"required"`
}

type ResolvePullRequestHeadSHAOutput struct {
	HeadSHA string `json:"head_sha"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResolvePullRequestHeadSHA(ctx context.Context, input *ResolvePullRequestHeadSHAInput) (*ResolvePullRequestHeadSHAOutput, error) {
	owner, repo, client, err := a.resolveGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		return nil, err
	}

	pr, _, err := client.PullRequests.Get(ctx, owner, repo, input.PRNumber)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return nil, nrErr
		}
		return nil, fmt.Errorf("unable to get pull request %d: %w", input.PRNumber, err)
	}

	sha := ""
	if pr != nil && pr.Head != nil {
		sha = pr.Head.GetSHA()
	}
	if sha == "" {
		return nil, vcserrors.NewGitRefNotFound(owner+"/"+repo, fmt.Sprintf("%d", input.PRNumber), nil)
	}
	return &ResolvePullRequestHeadSHAOutput{HeadSHA: sha}, nil
}
