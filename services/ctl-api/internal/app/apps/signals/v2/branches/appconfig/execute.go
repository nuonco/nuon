package appconfig

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if len(branch.Configs) == 0 {
		return fmt.Errorf("app branch has no config")
	}

	var vcsConfigID string
	if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else {
		return fmt.Errorf("app branch has no VCS config")
	}

	// Fetch the intermediate config from the repo
	appConfig, err := activities.AwaitFetchIntermediateConfig(ctx, activities.FetchIntermediateConfigRequest{
		VcsConfigID: vcsConfigID,
		CommitSHA:   s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch intermediate config: %w", err)
	}

	// Update app branch last synced commit
	if err := activities.AwaitUpdateAppBranchLastSyncedCommit(ctx, activities.UpdateAppBranchLastSyncedCommitRequest{
		AppBranchID: branch.ID,
		CommitSHA:   s.CommitSHA,
	}); err != nil {
		return fmt.Errorf("unable to update last synced commit: %w", err)
	}

	workflow.GetLogger(ctx).Info("intermediate config fetched",
		"app_branch_id", branch.ID,
		"commit_sha", s.CommitSHA,
		"config_version", appConfig.Version,
		"num_components", len(appConfig.Components))

	// TODO: sync the intermediate config via separate activities as a follow-up

	return nil
}
