package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FetchInstallSyncCommitInput struct {
	AppInstallConfigSyncID string `json:"app_install_config_sync_id" validate:"required"`
	VCSConfigID            string `json:"vcs_config_id" validate:"required"`
	CommitSHA              string `json:"commit_sha,omitempty"`
	Branch                 string `json:"branch,omitempty"`
}

type FetchInstallSyncCommitOutput struct {
	CommitID string `json:"commit_id"`
	SHA      string `json:"sha"`
	Message  string `json:"message"`
	Author   string `json:"author"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) FetchInstallSyncCommit(ctx context.Context, input *FetchInstallSyncCommitInput) (*FetchInstallSyncCommitOutput, error) {
	ref := input.Branch
	if input.CommitSHA != "" {
		ref = input.CommitSHA
	}

	var vcsCommit *app.VCSConnectionCommit

	fetchInput := &FetchCommitBySHAInput{
		VcsConfigID: input.VCSConfigID,
		SHA:         ref,
	}
	// Returned unwrapped: Temporal's failure converter type-switches on the
	// concrete error, so wrapping a non-retryable ApplicationError makes the
	// top-level failure retryable again.
	commit, err := a.FetchCommitBySHA(ctx, fetchInput)
	if err != nil {
		return nil, err
	}
	vcsCommit = commit

	if vcsCommit == nil {
		return nil, fmt.Errorf("unable to resolve commit data")
	}

	vcsCommit, err = a.createCommit(ctx, vcsCommit)
	if err != nil {
		return nil, fmt.Errorf("unable to create commit record: %w", err)
	}

	if err := a.db.WithContext(ctx).
		Model(&app.AppInstallConfigSync{}).
		Where("id = ?", input.AppInstallConfigSyncID).
		Update("vcs_connection_commit_id", vcsCommit.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to link commit to sync: %w", err)
	}

	return &FetchInstallSyncCommitOutput{
		CommitID: vcsCommit.ID,
		SHA:      vcsCommit.SHA,
		Message:  vcsCommit.Message,
		Author:   vcsCommit.AuthorName,
	}, nil
}
