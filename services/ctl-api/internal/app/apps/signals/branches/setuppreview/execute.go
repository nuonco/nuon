package setuppreview

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

const commitOnlyPreviewStatusesVersion = "app-branch-commit-only-preview-statuses-v1"

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if !run.IsPreview() {
		logger.Info("not a preview run, skipping setup")
		return nil
	}
	if workflow.GetVersion(ctx, commitOnlyPreviewStatusesVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return s.executeLegacy(ctx, logger, run)
	}

	// Statuses only need a commit; comments are what need a PR to live on.
	if !run.PreviewGitHubSetStatuses() && (!run.PreviewGitHubComment() || run.PRNumber == nil) {
		logger.Info("preview GitHub integration disabled, skipping setup")
		return nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	var vcsConfigID string
	if cfg := run.AppBranchConfig.ConnectedGithubVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if cfg := run.AppBranchConfig.PublicGitVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	}
	if vcsConfigID == "" {
		logger.Info("no VCS config found, skipping GitHub integration")
		return nil
	}

	if run.HeadSHA != "" && run.PreviewGitHubSetStatuses() {
		_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			State:       "pending",
			Description: "Preview starting...",
			AppBranchID: s.AppBranchID,
			RunID:       s.RunID,
			Preview:     true,
			PreviewMode: run.PreviewMode(),
		})
	}

	if !run.PreviewGitHubComment() || run.PRNumber == nil {
		logger.Info("preview setup complete", "run_id", s.RunID)
		return nil
	}

	prevComment, err := activities.AwaitFindPreviousRunCommentID(ctx, &activities.FindPreviousRunCommentIDInput{
		AppBranchID: s.AppBranchID,
		PRNumber:    *run.PRNumber,
	})
	if err != nil {
		logger.Warn("unable to find previous run comment, will create new", "error", err)
	}

	var existingCommentID *int64
	if prevComment != nil {
		existingCommentID = prevComment.CommentID
	}

	commentContext, _ := activities.AwaitGetPreviewCommentContext(ctx, &activities.GetPreviewCommentContextInput{
		RunID: s.RunID,
	})
	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		OrgName:    branch.Org.Name,
		AppName:    branch.App.Name,
		BranchName: branch.Name,
		RunID:      s.RunID,
		Status:     activities.PRCommentStatusPending,
		Mode:       run.PreviewMode(),
		RunURL:     previewRunURL(commentContext),
	})

	commentResult, err := activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		ExistingCommentID: existingCommentID,
		Body:              commentBody,
	})
	if err != nil {
		logger.Warn("unable to post PR comment", "error", err)
		return nil
	}

	if commentResult != nil && commentResult.CommentID != 0 {
		_ = activities.AwaitUpdateAppBranchRunGithubComment(ctx, &activities.UpdateAppBranchRunGithubCommentInput{
			RunID:     s.RunID,
			CommentID: commentResult.CommentID,
		})
		logger.Info("preview comment posted", "comment_id", commentResult.CommentID)
	}

	logger.Info("preview setup complete",
		"run_id", s.RunID,
		"pr_number", *run.PRNumber)

	return nil
}

func (s *Signal) executeLegacy(ctx workflow.Context, logger log.Logger, run *app.AppBranchRun) error {
	if run.PRNumber == nil {
		logger.Info("no PR number on preview run, skipping GitHub integration")
		return nil
	}
	if !run.PreviewGitHubSetStatuses() && !run.PreviewGitHubComment() {
		logger.Info("preview GitHub integration disabled, skipping setup")
		return nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	var vcsConfigID string
	if len(branch.Configs) > 0 {
		if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		}
	}
	if vcsConfigID == "" {
		return nil
	}

	prevComment, err := activities.AwaitFindPreviousRunCommentID(ctx, &activities.FindPreviousRunCommentIDInput{
		AppBranchID: s.AppBranchID,
		PRNumber:    *run.PRNumber,
	})
	if err != nil {
		logger.Warn("unable to find previous run comment, will create new", "error", err)
	}
	var existingCommentID *int64
	if prevComment != nil {
		existingCommentID = prevComment.CommentID
	}

	if run.HeadSHA != "" && run.PreviewGitHubSetStatuses() {
		_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			State:       "pending",
			Description: "Preview starting...",
			AppBranchID: s.AppBranchID,
			RunID:       s.RunID,
			Preview:     true,
			PreviewMode: run.PreviewMode(),
		})
	}
	if !run.PreviewGitHubComment() {
		return nil
	}

	commentContext, _ := activities.AwaitGetPreviewCommentContext(ctx, &activities.GetPreviewCommentContextInput{
		RunID: s.RunID,
	})
	commentResult, err := activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		ExistingCommentID: existingCommentID,
		Body: activities.BuildPRCommentBody(&activities.PRCommentParams{
			OrgName:    branch.Org.Name,
			AppName:    branch.App.Name,
			BranchName: branch.Name,
			RunID:      s.RunID,
			Status:     activities.PRCommentStatusPending,
			Mode:       run.PreviewMode(),
			RunURL:     previewRunURL(commentContext),
		}),
	})
	if err != nil {
		logger.Warn("unable to post PR comment", "error", err)
		return nil
	}
	if commentResult != nil && commentResult.CommentID != 0 {
		_ = activities.AwaitUpdateAppBranchRunGithubComment(ctx, &activities.UpdateAppBranchRunGithubCommentInput{
			RunID:     s.RunID,
			CommentID: commentResult.CommentID,
		})
	}

	return nil
}

func previewRunURL(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.RunURL
}
