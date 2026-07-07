package activities

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
)

type SetGithubCommitStatusInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
	State       string `json:"state" validate:"required"`
	Context     string `json:"context" validate:"required"`
	Description string `json:"description"`
	TargetURL   string `json:"target_url,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) SetGithubCommitStatus(ctx context.Context, input *SetGithubCommitStatusInput) error {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping commit status: no authenticated GitHub client available")
		return nil
	}

	status := &github.RepoStatus{
		State:       &input.State,
		Context:     &input.Context,
		Description: &input.Description,
	}
	if input.TargetURL != "" {
		status.TargetURL = &input.TargetURL
	}

	_, _, err = client.Repositories.CreateStatus(ctx, owner, repo, input.CommitSHA, status)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return nrErr
		}
		return fmt.Errorf("unable to create commit status: %w", err)
	}

	return nil
}
