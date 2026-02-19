package clonerepo

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

	result, err := activities.AwaitCloneRepo(ctx, activities.CloneRepoRequest{
		VcsConfigID: vcsConfigID,
		CommitSHA:   s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	workflow.GetLogger(ctx).Info("repo cloned successfully",
		"app_branch_id", branch.ID,
		"commit_sha", s.CommitSHA,
		"workspace_id", result.WorkspaceID,
		"source_dir", result.SourceDir)

	return nil
}
