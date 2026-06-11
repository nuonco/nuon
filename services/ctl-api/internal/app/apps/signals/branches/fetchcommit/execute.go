package fetchcommit

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	// Get the app branch
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// Check if branch has configs
	if len(branch.Configs) == 0 {
		logger.Info("no configs found for app branch", "app_branch_id", branch.ID)
		return nil
	}

	cfg := branch.Configs[0]

	// Determine the VCS config ID to use
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

	// Fetch the latest commit from GitHub (no DB interaction)
	vcsCommit, err := activities.AwaitFetchLatestCommitByVcsConfigID(ctx, vcsConfigID)
	if err != nil {
		return fmt.Errorf("unable to fetch latest commit: %w", err)
	}

	// Create the commit record in the database
	vcsCommit, err = activities.AwaitCreateCommitByVcsCommit(ctx, vcsCommit)
	if err != nil {
		return fmt.Errorf("unable to create commit: %w", err)
	}

	logger.Info("fetched commit from VCS",
		"app_branch_id", branch.ID,
		"commit_sha", vcsCommit.SHA,
		"author", vcsCommit.AuthorName,
		"vcs_commit_id", vcsCommit.ID)

	// Update the app branch run with the VCS commit ID
	err = activities.AwaitUpdateAppBranchRunVCSCommit(ctx, activities.UpdateAppBranchRunVCSCommitRequest{
		RunID:       s.RunID,
		VcsCommitID: vcsCommit.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to update run with VCS commit: %w", err)
	}

	// Update step metadata with commit details for the UI
	if s.StepID != "" {
		meta := map[string]any{
			"commit_sha":     vcsCommit.SHA,
			"commit_message": vcsCommit.Message,
			"author_name":    vcsCommit.AuthorName,
			"author_email":   vcsCommit.AuthorEmail,
		}

		var vcsConfigID string
		if cfg.ConnectedGithubVCSConfig != nil {
			meta["repo"] = cfg.ConnectedGithubVCSConfig.Repo
			meta["branch"] = cfg.ConnectedGithubVCSConfig.Branch
			vcsConfigID = cfg.ConnectedGithubVCSConfig.ID
		}
		if cfg.PublicGitVCSConfig != nil {
			meta["repo"] = cfg.PublicGitVCSConfig.Repo
			meta["branch"] = cfg.PublicGitVCSConfig.Branch
			vcsConfigID = cfg.PublicGitVCSConfig.ID
		}

		// Get the run's base branch info
		run, runErr := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
		if runErr == nil && run.BaseBranch != "" {
			meta["base_branch"] = run.BaseBranch
		}

		// Best-effort: fetch PR info for the branch
		if vcsConfigID != "" {
			branchName := ""
			if b, ok := meta["branch"].(string); ok {
				branchName = b
			}
			if branchName != "" {
				prInfo, prErr := activities.AwaitFetchCommitPRInfo(ctx, &activities.FetchCommitPRInfoInput{
					VcsConfigID: vcsConfigID,
					Branch:      branchName,
				})
				if prErr == nil && prInfo != nil {
					meta["pr_number"] = prInfo.PRNumber
					meta["pr_title"] = prInfo.PRTitle
					meta["pr_status"] = prInfo.PRStatus
					meta["pr_reviewer_count"] = prInfo.ReviewerCount
					meta["pr_url"] = prInfo.PRURL
				}
			}

			// Best-effort: fetch diff stats
			baseBranch := ""
			if b, ok := meta["base_branch"].(string); ok {
				baseBranch = b
			}
			diffStats, diffErr := activities.AwaitFetchCommitDiffStats(ctx, &activities.FetchCommitDiffStatsInput{
				VcsConfigID: vcsConfigID,
				CommitSHA:   vcsCommit.SHA,
				BaseBranch:  baseBranch,
			})
			if diffErr == nil && diffStats != nil {
				meta["files_changed"] = diffStats.FilesChanged
				meta["additions"] = diffStats.Additions
				meta["deletions"] = diffStats.Deletions
				meta["changed_files"] = diffStats.ChangedFiles
			}
		}

		_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.StepID,
			Status: app.CompositeStatus{
				Status:                 app.StatusSuccess,
				StatusHumanDescription: fmt.Sprintf("fetched commit %s", vcsCommit.SHA[:8]),
				Metadata:               meta,
			},
		})
	}

	logger.Info("successfully fetched and stored commit",
		"run_id", s.RunID,
		"app_branch_id", branch.ID,
		"commit_sha", vcsCommit.SHA)

	return nil
}
