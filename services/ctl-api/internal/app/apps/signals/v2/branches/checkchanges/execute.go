package checkchanges

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// No configs — nothing to check
	if len(branch.Configs) == 0 {
		logger.Info("no configs found for app branch", "app_branch_id", branch.ID)
		return nil
	}

	cfg := branch.Configs[0]

	// Determine the VCS config ID to check
	var vcsConfigID string
	switch {
	case cfg.ConnectedGithubVCSConfig != nil:
		vcsConfigID = cfg.ConnectedGithubVCSConfig.ID
	case cfg.PublicGitVCSConfig != nil:
		vcsConfigID = cfg.PublicGitVCSConfig.ID
	default:
		logger.Info("no VCS config found for app branch", "app_branch_id", branch.ID)
		return nil
	}

	// Fetch latest commit from VCS
	latestCommit, err := activities.AwaitGetLatestCommitFromVCSByVcsConfigID(ctx, vcsConfigID)
	if err != nil {
		return fmt.Errorf("unable to get latest commit: %w", err)
	}

	// Get the commit SHA from the most recent successful run for comparison
	lastRunCommit, err := activities.AwaitGetLatestAppBranchRunCommitSHAByAppBranchID(ctx, branch.ID)
	if err != nil {
		return fmt.Errorf("unable to get latest run commit: %w", err)
	}

	// Compare with last successful run's commit
	if latestCommit == lastRunCommit {
		logger.Info("no changes detected",
			"app_branch_id", branch.ID,
			"commit", latestCommit)
		return nil
	}

	logger.Info("changes detected",
		"app_branch_id", branch.ID,
		"latest_commit", latestCommit,
		"last_run_commit", lastRunCommit)

	// TODO: Enqueue update-app-config or run signal when ready

	return nil
}
